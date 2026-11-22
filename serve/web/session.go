package web

import (
	//安全随机数
	"crypto/rand"
	"fmt"
	"time"

	//hex 十六进制与字节之间的转换
	"encoding/hex"
	"encoding/json"
	"net"
	"sync"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

// 每个会话最多保留的消息条数
const maxMsgs = 200

type storeMsg struct {
	Meg message.Message
	At  time.Time
}

type Session struct {
	Conn     net.Conn
	UserId   int
	UserName string
	//浏览器http与tcp连接读取时间不一致,需要预存信息
	Msgs []storeMsg
	mu   sync.Mutex

	Users map[int]*message.User

	Pending  map[string]*pendingMes
	Reported map[string]bool
}

type pendingMes struct {
	Data   []byte
	SentAt time.Time
	Count  int
}

var (
	sessMu   sync.RWMutex
	sessions = make(map[string]*Session)
)

func (ss *Session) addPending(mesId string, data []byte) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.Pending[mesId] = &pendingMes{
		Data:   data,
		SentAt: time.Now(),
		Count:  0,
	}
}

func (ss *Session) removePending(mesId string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	delete(ss.Pending, mesId)
}

func (ss *Session) snapshotIDs() []string {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	var list []string
	for i := range ss.Pending {
		list = append(list, i)
	}
	return list
}

const (
	notExist   = 1 //不存在
	maxout     = 2 //超过最大重发数
	timenotout = 3 //没超时
	timeout    = 4 //超时
)

func (ss *Session) takePending(mesId string, maxCount int, maxTime time.Duration) ([]byte, bool, int) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	var state int
	//判断id是否存在待办
	pendingMes, ok := ss.Pending[mesId]
	if !ok {
		state = notExist
		return nil, false, state
	}

	//判断是否超过最大重发数
	if pendingMes.Count >= maxCount {
		state = maxout
		return nil, false, state
	}

	//判断是否超时(超过3s为超时)
	//还没到 3 秒 → return nil, false（不重发）；
	//超过 3 秒 → 继续往下走，计数、重发
	if time.Since(pendingMes.SentAt) < maxTime {
		//没超时不重发
		state = timenotout
		return nil, false, state
	}

	pendingMes.Count++
	pendingMes.SentAt = time.Now()
	data := append([]byte(nil), pendingMes.Data...)
	//超时重发
	state = timeout
	return data, true, state

}

func (ss *Session) markReported(mesId string) bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.Reported[mesId] {
		return false
	}

	ss.Reported[mesId] = true
	return true
}

func (ss *Session) delReported(mesId string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	delete(ss.Reported, mesId)
}

func addSession(sid string, s *Session) {
	sessMu.Lock()
	defer sessMu.Unlock()

	sessions[sid] = s

}

func getSession(sid string) (*Session, bool) {
	sessMu.RLock()
	defer sessMu.RUnlock()

	sess, ok := sessions[sid]
	return sess, ok
}

// 删除会话
func delSession(sid string) {
	sessMu.Lock()
	s, ok := sessions[sid]
	delete(sessions, sid)
	sessMu.Unlock()

	if ok {
		s.Conn.Close()
	}

}

// 生成一个随机sessionID,给cookie充当用户身份验证
func newSessionID() string {
	b := make([]byte, 16)
	//把随机数塞满b
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	//cookie只收文本，随机[]byte有可能包含空格、特殊符号等,需要转码

	//十六进制处理后成字符串
	//hex只是其中一种编码方式，还有base64 base32等等
	//hex是文本编码中最省事的一种
	return hex.EncodeToString(b)

}

// 浏览器的心跳保活
// web.go->serve
func (ss *Session) sendping() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		data, err := buildpingMes()
		if err != nil {
			fmt.Println("组装pingmes失败，err=", err)
			return
		}

		tf := utils.Transfer{
			Conn: ss.Conn,
		}

		err = tf.WritePkg(data)
		if err != nil {
			fmt.Println("writePkg data err=", err)
			return
		}
	}

}

// 组装pingmes信息
func buildpingMes() (data []byte, err error) {
	var mes message.Message
	mes.Type = message.PingMesType

	var pingmes message.PingMes
	pingmes.Timestamp = time.Now().UnixMilli()

	data, err = json.Marshal(pingmes)
	if err != nil {
		fmt.Println("marshal err=", err)
		return nil, err
	}
	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal err=", err)
		return nil, err
	}

	return data, err

}

// 持续监听ackMes 进行重发信息
func (ss *Session) retryAckMes() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	tf := utils.Transfer{
		Conn: ss.Conn,
	}
	for range ticker.C {
		for _, v := range ss.snapshotIDs() {
			data, ok, state := ss.takePending(v, 3, 3*time.Second)

			if !ok {
				//不在待办列表、超时、超次数,就直接删除
				switch state {
				case notExist:
				case maxout:
					ss.removePending(v)
					ss.delReported(v)
					fmt.Printf("消息 %s 重发 3 次仍未确认，已放弃\n", v)
				case timenotout:
				default:
					fmt.Println("retryAckMes未知错误")
				}
			} else if ok && state == timeout {
				//进行重发信息
				err := tf.WritePkg(data)
				if err != nil {
					fmt.Println("writePkg retryMes err=", err)
					ss.removePending(v)
					ss.delReported(v)
					return
				}
			}
		}
	}

}

// 读服务端发送的信息，并存入Msgs
func (ss *Session) readloop() {
	tf := utils.Transfer{
		Conn:    ss.Conn,
		Timeout: 30 * time.Second,
	}

	for {
		show := true //默认存进Msgs

		mes, err := tf.ReadPkg()
		if err != nil {
			fmt.Println("readPkg err=", err)
			return
		}

		//跳过心跳的返回信息
		if mes.Type == message.PongMesType {
			continue
		}

		//收到服务器转发的私聊时，发送回执信息
		if mes.Type == message.SmsToUserTransferMesType {
			var toUsermes message.SmsToUserTransferMes
			if err := json.Unmarshal([]byte(mes.Data), &toUsermes); err == nil {
				var m message.Message
				m.Type = message.SmsAckMesType

				ackmes := message.SmsAckMes{
					FromUserId:   ss.UserId,
					FromUserName: ss.UserName,

					ToUserId:   toUsermes.UserId,
					ToUserName: toUsermes.UserName,

					MessageId: toUsermes.MessageId,
				}

				data, _ := json.Marshal(ackmes)

				m.Data = string(data)

				data, _ = json.Marshal(m)

				tf := utils.Transfer{
					Conn: ss.Conn,
				}
				if err := tf.WritePkg(data); err != nil {
					fmt.Println("send ack err=", err)
				}
			}
		}

		//收到私聊响应信息时
		if mes.Type == message.SmsToUserResMesType {
			var res message.SmsToUserResMes
			if err := json.Unmarshal([]byte(mes.Data), &res); err == nil {
				if res.Code != 200 {
					ss.removePending(res.MessageId)
					ss.delReported(res.MessageId)
				} else if !ss.markReported(res.MessageId) {
					show = false // 第二次及以后的 200，不显示
				}
			}
		}

		//收到回执信息时
		if mes.Type == message.SmsAckTrsMesType {
			var ack message.SmsAckTrsMes
			if err := json.Unmarshal([]byte(mes.Data), &ack); err == nil {
				ss.removePending(ack.MessageId)
				ss.delReported(ack.MessageId)
			}
		}

		//更新在线表
		var online *message.User
		var offlineId int
		if mes.Type == message.NotifyUserStatusMesType {
			var m message.NotifyUserStatusMes
			if err := json.Unmarshal([]byte(mes.Data), &m); err != nil {
				continue
			}

			//在线增加
			if m.Status == message.UserOnline {
				online = &message.User{
					UserId:     m.UserId,
					UserName:   m.UserName,
					UserStatus: m.Status,
				}
			} else { //离线清除
				offlineId = m.UserId
			}

		}

		ss.mu.Lock()

		if online != nil { //增加在线
			ss.Users[online.UserId] = online
		} else if offlineId != 0 { //清理离线
			delete(ss.Users, offlineId)
		}

		//消息列表
		if show {
			ss.Msgs = append(ss.Msgs, storeMsg{
				Meg: mes,
				At:  time.Now(),
			})

			//只保留最近 maxMsgs 条，避免长时间挂机越积越多
			if len(ss.Msgs) > maxMsgs {
				ss.Msgs = ss.Msgs[len(ss.Msgs)-maxMsgs:]
			}

		}

		ss.mu.Unlock()
	}
}

// 复制一份消息列表
func (ss *Session) snapshot() []storeMsg {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	list := make([]storeMsg, len(ss.Msgs))
	copy(list, ss.Msgs)
	return list
}

// 复制一份在线用户列表
func (ss *Session) snapshotUsers() []message.User {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	list := make([]message.User, 0, len(ss.Users))

	for _, v := range ss.Users {
		list = append(list, *v)
	}

	return list
}
