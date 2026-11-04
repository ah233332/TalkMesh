package process

import (
	"encoding/json"
	"fmt"

	"github.com/ah233332/TalkMesh/common/message"
)

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

// 解析服务端转发来的私聊消息，并把发送者、内容打印到控制台。
func outputToUserMes(mes *message.Message) (err error) {
	var ToUserTransferMes message.SmsToUserTransferMes
	err = json.Unmarshal([]byte(mes.Data), &ToUserTransferMes)
	if err != nil {
		fmt.Println("unmarshal mes data err=", err)
		return
	}
	fmt.Printf("%s(%d)给你私发了一条消息:%s\n", ToUserTransferMes.UserName, ToUserTransferMes.UserId, ToUserTransferMes.Content)
	return
}
