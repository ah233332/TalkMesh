package process

import (
	"encoding/json"
	"fmt"

	"github.com/ah233332/TalkMesh/common/message"
)

// 解析服务端转发来的群发消息，并把发送者、内容打印到控制台。
func outputGroupMes(mes *message.Message) (err error) {
	var smsTransferMes message.SmsTransferMes
	err = json.Unmarshal([]byte(mes.Data), &smsTransferMes)
	if err != nil {
		fmt.Println("unmarshal mes data err=", err)
		return
	}

	fmt.Printf("%s(%d)发送了一条消息:%s\n", smsTransferMes.UserName, smsTransferMes.UserId, smsTransferMes.Content)
	return
}
