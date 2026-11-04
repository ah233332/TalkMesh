package process

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
	"github.com/ah233332/TalkMesh/serve/model"
)

type SmsProcess struct {
	Conn net.Conn
}

// 处理群聊消息的核心函数方法
func (sp *SmsProcess) ServeGroupMes(mes *message.Message) (err error) {
	//1.先把消息解出来
	var smsMes message.SmsMessage
	err = json.Unmarshal([]byte(mes.Data), &smsMes)
	if err != nil {
		fmt.Println("unmarshal mes data err=", err)
		return
	}

	mes.Type = message.SmsTransferMesType

	data, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal mes err=", err)
		return
	}
	for _, up := range userMgr.Snapshot() {

		errSend := sp.sendMes(up.Conn, data)
		if errSend != nil {
			fmt.Printf("发送给用户:%d 发送失败\n", up.UserId)
			continue
		}
	}

	return
}

// 主要功能用来给在线用户发送信息
func (sp *SmsProcess) sendMes(conn net.Conn, data []byte) (err error) {

	tf := utils.Transfer{
		Conn: conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("sendGroupMes writePkg data err=", err)
		return
	}
	return
}

// 服务私聊的核心代码，转发信息+返回信息
func (sp *SmsProcess) ServeToUserMes(mes *message.Message) (err error) {
	//转发信息
	err = sp.transferToUserMes(mes)

	//封装返回客户端的信息
	mes.Type = message.SmsToUserResMesType
	var toUserResMes message.SmsToUserResMes
	if err != nil {
		switch err {
		case model.ERROR_USER_NOTONLINE:
			toUserResMes.Code = 400
			toUserResMes.Error = err.Error()
		default:
			toUserResMes.Code = 505
			toUserResMes.Error = err.Error()
		}
	} else {
		toUserResMes.Code = 200
	}

	data, err := json.Marshal(toUserResMes)
	if err != nil {
		fmt.Println("marshal toUserResMes err=", err)
		return err
	}
	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal mes err=", err)
		return err
	}
	//此时data为可直接返回的信息

	tf := utils.Transfer{
		Conn: sp.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("writePkg mes err=", err)
		return err
	}
	return
}

// 转发私信的函数方法
func (sp *SmsProcess) transferToUserMes(mes *message.Message) (err error) {
	var toUserMes message.SmsToUserMes
	err = json.Unmarshal([]byte(mes.Data), &toUserMes)
	if err != nil {
		fmt.Println("unmarshal mes data err=", err)
		return
	}

	up, err := userMgr.GetOnlineUserById(toUserMes.ToUserId)
	if err != nil {
		err = model.ERROR_USER_NOTONLINE
		return
	}

	mes.Type = message.SmsToUserTransferMesType

	data, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal mes err=", err)
		return err
	}

	tf := utils.Transfer{
		Conn: up.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("sendGroupMes writePkg data err=", err)
		return
	}
	return
}
