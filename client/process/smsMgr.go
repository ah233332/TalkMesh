package process

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

const maxseen = 1000

var (
	reportMu sync.Mutex
	reported = make(map[string]bool) //待办信息

	seenMu   sync.Mutex
	seen     = make(map[string]bool)      //去重信息
	seenlist = make([]string, 0, maxseen) //记录信息顺序，用来淘汰
)

func markReported(mesId string) bool {
	reportMu.Lock()
	defer reportMu.Unlock()

	if reported[mesId] {
		return false
	}

	reported[mesId] = true
	return true

}

func delReported(mesId string) {
	reportMu.Lock()
	defer reportMu.Unlock()

	delete(reported, mesId)

}

// 解析服务端转发来的群发消息，并把发送者、内容打印到控制台。
func outputGroupMes(mes *message.Message) (err error) {
	var groupTransferMes message.SmsTransferMes
	err = json.Unmarshal([]byte(mes.Data), &groupTransferMes)
	if err != nil {
		fmt.Println("unmarshal mes data err=", err)
		return
	}

	fmt.Printf("%s(%d)群发送了一条消息:%s\n", groupTransferMes.UserName, groupTransferMes.UserId, groupTransferMes.Content)
	return
}

// 服务服务端转发来的私聊消息
func serveToUserTrsMes(mes *message.Message) (err error) {
	var ToUserTransferMes message.SmsToUserTransferMes
	err = json.Unmarshal([]byte(mes.Data), &ToUserTransferMes)
	if err != nil {
		fmt.Println("unmarshal mes data err=", err)
		return
	}

	//取出信息的mesID
	mesId := ToUserTransferMes.MessageId

	//加锁查去重表
	seenMu.Lock()
	ok := seen[mesId]
	if ok {
		//存在说明打印输出过了，不重复打印
		seenMu.Unlock()

		//发送ACK回执信息
		err = sendAckMes(ToUserTransferMes)
		if err != nil {
			fmt.Println("sendAckmes err=", err)
			return
		}
		return
	} else {
		//不存在则登记去重表，打印内容
		seen[mesId] = true
		seenlist = append(seenlist, mesId)

		//删除旧的消息内容
		if len(seenlist) > 1000 {
			oldest := seenlist[0]
			seenlist = seenlist[1:]
			delete(seen, oldest)
		}

		//防止内存占用越滚越多
		if cap(seenlist) > 2*len(seenlist) {
			// 底层数组浪费超过一半，压一次
			compact := make([]string, len(seenlist))
			copy(compact, seenlist)
			seenlist = compact
		}

		seenMu.Unlock()

		//打印私信内容
		outputTouserMes(ToUserTransferMes)

		//发送ACK回执信息
		err = sendAckMes(ToUserTransferMes)
		if err != nil {
			fmt.Println("sendAckmes err=", err)
			return
		}
		return
	}

}

// 打印显示出私信内容
func outputTouserMes(ToUserTrsMes message.SmsToUserTransferMes) {
	fmt.Printf("%s(%d)给你私发了一条消息:%s\n", ToUserTrsMes.UserName, ToUserTrsMes.UserId, ToUserTrsMes.Content)
}

// 用来发送私聊回执信息
func sendAckMes(ToUserTrsMes message.SmsToUserTransferMes) (err error) {
	//插入回执信息函数
	data := CreatAckMes(ToUserTrsMes.MessageId, curUser.UserId, curUser.UserName, ToUserTrsMes.UserId, ToUserTrsMes.UserName)
	tf := utils.Transfer{
		Conn: curUser.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("writePkg data err=", err)
		return
	}
	return
}
