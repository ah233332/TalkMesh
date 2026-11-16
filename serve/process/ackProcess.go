package process

import (
	"encoding/json"
	"fmt"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

type AckProcess struct {
}

// 用来处理回执信息的
func (ap *AckProcess) ServeAckMes(mes *message.Message) {
	var ackmes message.SmsAckMes
	err := json.Unmarshal([]byte(mes.Data), &ackmes)
	if err != nil {
		fmt.Println("unmarshal err=", err)
		return
	}

	up, err := userMgr.GetOnlineUserById(ackmes.ToUserId)
	if err != nil {
		fmt.Println("getonlineUser err=", err)
		return
	}

	tf := utils.Transfer{
		Conn: up.Conn,
	}
	data := creatAckResMes(ackmes)
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("writepkg data err=", err)
		return
	}
}

func creatAckResMes(ackmes message.SmsAckMes) []byte {
	var mes message.Message
	mes.Type = message.SmsAckTrsMesType

	var acktfsmes message.SmsAckTrsMes
	acktfsmes.FromUserId = ackmes.FromUserId
	acktfsmes.ToUserId = ackmes.ToUserId

	data, err := json.Marshal(acktfsmes)
	if err != nil {
		fmt.Println("marshal acktfsmes err=", err)
		return nil
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal mes err=", err)
		return nil
	}
	return data
}
