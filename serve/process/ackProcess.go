package process

import (
	"encoding/json"
	"fmt"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

type AckProcess struct {
	UserId   int
	UserName string
}

// 用来处理回执信息的
func (ap *AckProcess) ServeAckMes(mes *message.Message) {
	var ackmes message.SmsAckMes
	err := json.Unmarshal([]byte(mes.Data), &ackmes)
	if err != nil {
		fmt.Println("unmarshal err=", err)
		return
	}

	//安全校验
	if ap.UserId == 0 {
		fmt.Println("未登录连接发送回执")
		return
	}

	toup, err := userMgr.GetOnlineUserById(ackmes.ToUserId)
	if err != nil {
		fmt.Println("getonlineUser err=", err)
		return
	}

	tf := utils.Transfer{
		Conn: toup.Conn,
	}
	data := ap.creatAckTrsMes(ackmes, toup)
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("writepkg data err=", err)
		return
	}
}

func (ap *AckProcess) creatAckTrsMes(ackmes message.SmsAckMes, toup *UserProcess) []byte {
	acktfsmes := message.SmsAckTrsMes{
		FromUserId:   ap.UserId,
		FromUserName: ap.UserName,

		ToUserId:   toup.UserId,
		ToUserName: toup.UserName,

		MessageId: ackmes.MessageId,
	}

	data, err := utils.BuildPkg(message.SmsAckTrsMesType, acktfsmes)
	if err != nil {
		fmt.Println("bulid smsacktrsmes err=", err)
		return nil
	}
	return data
}
