package process

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

type PingProcess struct {
}

// 客户端定时发送ping给服务端
func (pp *PingProcess) SendPingMes(conn net.Conn) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		data, err := pp.pingMes()
		if err != nil {
			fmt.Println("组装pingMes失败，err=", err)
			return
		}

		tf := utils.Transfer{Conn: conn}

		err = tf.WritePkg(data)
		if err != nil {
			fmt.Println("writePkg pingMes err=", err)
			return
		}
	}

}

// 组装一个可直接发送的ping信息
func (pp *PingProcess) pingMes() ([]byte, error) {
	var mes message.Message
	mes.Type = message.PingMesType

	var pingMes message.PingMes
	pingMes.Timestamp = time.Now().UnixMilli()

	data, err := json.Marshal(pingMes)
	if err != nil {
		fmt.Println("marshal pingmes err=", err)
		return nil, err
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal mes err=", err)
		return nil, err
	}
	//此时data就为可直接发送的信息
	return data, nil

}
