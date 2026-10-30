package process

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

type SmsProcess struct{}

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
	for _, up := range userMgr.onlineUsers {

		err = sp.sendGroupMes(up.Conn, data)
		if err != nil {
			fmt.Println("该用户发送失败", up.UserId)
			continue
		}
	}

	return
}

// 用来服务sendGroupMes函数的方法
// 主要功能用来给在线用户发送信息
func (sp *SmsProcess) sendGroupMes(conn net.Conn, data []byte) (err error) {

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
