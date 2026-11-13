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
}

var (
	sessMu   sync.RWMutex
	sessions = make(map[string]*Session)
)

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

// 读服务端发送的信息，并存入Msgs
func (ss *Session) readloop() {
	tf := utils.Transfer{
		Conn:    ss.Conn,
		Timeout: 30 * time.Second,
	}

	for {
		mes, err := tf.ReadPkg()
		if err != nil {
			fmt.Println("readPkg err=", err)
			return
		}

		//跳过心跳的返回信息
		if mes.Type == message.PongMesType {
			continue
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
		ss.Msgs = append(ss.Msgs, storeMsg{
			Meg: mes,
			At:  time.Now(),
		})

		//只保留最近 maxMsgs 条，避免长时间挂机越积越多
		if len(ss.Msgs) > maxMsgs {
			ss.Msgs = ss.Msgs[len(ss.Msgs)-maxMsgs:]
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
