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
	mes.Type = message.SmsGPMessageType

	var smsMes = message.SmsGPMessage{
		Content:    content,
		UserId:     sp.UserId,
		UserName:   sp.UserName,
		UserStatus: sp.UserStatus,

		MessageId: utils.CreatMesId(sp.UserId),
	}

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

	//消息id
	messageID := utils.CreatMesId(sp.UserId)

	smstoUserMes := message.SmsToUserMes{
		ToUserId:   toUserId,
		Content:    content,
		UserId:     sp.UserId,
		UserName:   sp.UserName,
		UserStatus: sp.UserStatus,

		MessageId: messageID,
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

	//把消息加入待办列表
	addPending(messageID, data)

	tf := utils.Transfer{
		Conn: sp.conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("write SmsMes err=", err)
		//发送失败就删除列表
		removePending(messageID)
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
		if markReported(toUserResMes.MessageId) {
			fmt.Println("成功发送私聊信息")
		}
	case 201:
		removePending(toUserResMes.MessageId)
		fmt.Println("目标用户不在线，成功发送离线消息")
	case 400:
		removePending(toUserResMes.MessageId)
		fmt.Println("该用户不在线")
	case 401:
		removePending(toUserResMes.MessageId)
		fmt.Println("该用户未登录")
	case 402:
		removePending(toUserResMes.MessageId)
		fmt.Println("发送对象错误")
	case 403:
		removePending(toUserResMes.MessageId)
		fmt.Println("发送者身份不一致")
	case 405:
		removePending(toUserResMes.MessageId)
		fmt.Println("目标用户不存在")
	case 505:
		removePending(toUserResMes.MessageId)
		fmt.Println("未知错误")
	default:
		removePending(toUserResMes.MessageId)
		fmt.Println("???")
	}
	return
}
