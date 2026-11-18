package process

import (
	"encoding/json"
	"fmt"

	"github.com/ah233332/TalkMesh/common/message"
)

type AckProcess struct {
}

// formId B 谁收到的私信(自己)  toId A 回执给谁(私信发送者)
func CreatAckMes(formId int, fromName string, toId int, toName string) []byte {
	var mes message.Message
	mes.Type = message.SmsAckMesType

	ackmes := message.SmsAckMes{
		FromUserId:   formId,
		FromUserName: fromName,
		ToUserId:     toId,
		ToUserName:   toName,
	}

	data, err := json.Marshal(ackmes)
	if err != nil {
		fmt.Println("marshal ackmes err=", err)
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
