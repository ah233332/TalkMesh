package process

import (
	"encoding/json"
	"fmt"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

type SmsProcess struct {
}

// 群发消息的函数方法
func (sp *SmsProcess) SendGroupMes(content string) (err error) {
	//封装一个可发送的Mes信息
	var mes message.Message
	mes.Type = message.SmsMessageType

	var smsMes message.SmsMessage
	smsMes.Content = content
	smsMes.UserId = curUser.UserId         //id
	smsMes.UserName = curUser.UserName     //名称
	smsMes.UserStatus = curUser.UserStatus //状态

	data, err := json.Marshal(smsMes)
	if err != nil {
		fmt.Println("marshal smsMes err=", err)
		return
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal mes err=", err)
		return
	}
	//此时data就是可发送的消息
	//发送信息
	tf := utils.Transfer{
		Conn: curUser.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("write SmsMes err=", err)
		return
	}

	return
}
