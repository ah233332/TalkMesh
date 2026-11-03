package process

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

type PingProcess struct {
	Conn net.Conn
}

// 接收处理客户端发送的pingMes心跳，并返回pongMes信息
func (pp *PingProcess) ServePing(mes *message.Message) error {
	mes.Type = message.PongMesType
	var pingMes message.PingMes

	err := json.Unmarshal([]byte(mes.Data), &pingMes)
	if err != nil {
		fmt.Println("unmarshal mes data err=", err)
		return err
	}
	var pongMes message.PongMes
	pongMes.Timestamp = pingMes.Timestamp

	data, err := json.Marshal(pongMes)
	if err != nil {
		fmt.Println("marshal pongmes err=", err)
		return err
	}
	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal mes err=", err)
		return err
	}
	//此时data就是可直接发送的pongMes

	tf := utils.Transfer{
		Conn: pp.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("writePkg data err=", err)
		return err
	}

	return nil
}
