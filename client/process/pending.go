package process

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ah233332/TalkMesh/common/utils"
)

var (
	pendingMu sync.Mutex
	pending   = make(map[string]*pendingMes) //等待回执信息的表
)

type pendingMes struct {
	Data   []byte
	SentAt time.Time
	Count  int
}

const (
	notExist   = 1 //不存在
	maxout     = 2 //超过最大重发数
	timenotout = 3 //没超时
	timeout    = 4 //超时
)

func retryAckMes(conn net.Conn) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	tf := utils.Transfer{
		Conn: conn,
	}
	for range ticker.C {
		for _, v := range snapshotIDs() {
			data, ok, state := takeSpending(v, 3, 3*time.Second)

			if !ok {
				//不在待办列表、超时、超次数,就直接删除
				switch state {
				case notExist:
				case maxout:
					removePending(v)
					delReported(v)
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
					removePending(v)
					delReported(v)
					return
				}
			}
		}
	}

}

// 判断信息现在该重发吗？
// 该重发返回true  不该重发就返回false
func takeSpending(id string, maxCount int, maxTime time.Duration) (data []byte, ok bool, state int) {
	pendingMu.Lock()
	defer pendingMu.Unlock()

	//判断id是否存在待办
	pendingMes, ok := pending[id]
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
	data = append([]byte(nil), pendingMes.Data...)
	//超时重发
	state = timeout
	return data, true, state

}

func addPending(id string, data []byte) {
	pendingMu.Lock()
	defer pendingMu.Unlock()

	pending[id] = &pendingMes{
		Data:   data,
		SentAt: time.Now(),
		Count:  0,
	}
}

func removePending(id string) {
	pendingMu.Lock()
	defer pendingMu.Unlock()

	delete(pending, id)

}

func snapshotIDs() []string {
	pendingMu.Lock()
	defer pendingMu.Unlock()

	var list []string
	for k := range pending {
		list = append(list, k)
	}
	return list
}
