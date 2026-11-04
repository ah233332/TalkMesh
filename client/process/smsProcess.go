package process

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

type SmsProcess struct {
	message.User
	conn net.Conn
}

// 群发消息的函数方法
func (sp *SmsProcess) SendGroupMes(content string) (err error) {
	//封装一个可发送的Mes信息
	var mes message.Message
	mes.Type = message.SmsMessageType

	var smsMes message.SmsMessage
	smsMes.Content = content
	smsMes.UserId = sp.UserId         //id
	smsMes.UserName = sp.UserName     //名称
	smsMes.UserStatus = sp.UserStatus //状态

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
		Conn: sp.conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("write SmsMes err=", err)
		return
	}

	return
}

// 私发消息的函数方法
func (sp *SmsProcess) SendToUserMes(toUserId int, content string) {
	//封装一个可发送的私聊请求
	var mes message.Message
	mes.Type = message.SmsToUserMesType

	smstoUserMes := message.SmsToUserMes{
		ToUserId:   toUserId,
		Content:    content,
		UserId:     sp.UserId,
		UserName:   sp.UserName,
		UserStatus: sp.UserStatus,
	}

	data, err := json.Marshal(smstoUserMes)
	if err != nil {
		fmt.Println("marshal smstoUserMes err=", err)
		return
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal Mes err=", err)
		return
	}
	//data为可直接发送的私聊请求

	tf := utils.Transfer{
		Conn: sp.conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("write SmsMes err=", err)
		return
	}

}

// 处理私聊服务端响应返回的信息
func (sp *SmsProcess) ServeToUserResMes(mes *message.Message) (err error) {
	var toUserResMes message.SmsToUserResMes
	err = json.Unmarshal([]byte(mes.Data), &toUserResMes)
	if err != nil {
		fmt.Println("unmarshal mes data err=", err)
		return
	}

	switch toUserResMes.Code {
	case 200:
		fmt.Println("成功发送私聊信息")
	case 400:
		fmt.Println("该用户不在线")
	case 505:
		fmt.Println("未知错误")
	default:
		fmt.Println("???")
	}
	return
}
