package process

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
)

type AckProcess struct {
	Conn   net.Conn
	FromId int
	ToId   int
}

func (ap *AckProcess) ServeAckMes() {

}

// formId 谁收到的私信(自己)  toId 回执给谁(私信发送者)
func CreatAckMes(formId int, toId int) []byte {
	var mes message.Message
	mes.Type = message.SmsAckMesType

	ackmes := message.SmsAckMes{
		FromUserId: formId,
		ToUserId:   toId,
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
