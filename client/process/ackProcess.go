package process

import (
	"fmt"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

type AckProcess struct {
}

// formId B 谁收到的私信(自己)  toId A 回执给谁(私信发送者)
func CreatAckMes(mesID string, formId int, fromName string, toId int, toName string) []byte {
	ackmes := message.SmsAckMes{
		FromUserId:   formId,
		FromUserName: fromName,
		ToUserId:     toId,
		ToUserName:   toName,

		MessageId: mesID,
	}

	data, err := utils.BuildPkg(message.SmsAckMesType, ackmes)
	if err != nil {
		fmt.Println("bulidpkg ackmes err=", err)
		return nil
	}

	return data

}
