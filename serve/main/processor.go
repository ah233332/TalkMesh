package main

//中转路由功能

import (
	"fmt"
	"io"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
	"github.com/ah233332/TalkMesh/serve/process"
)

type Processor struct {
	Conn net.Conn
}

// 路由功能,根据客户端发送的消息种类不同，决定调用哪个函数处理
func (pr *Processor) ServeProcessMes(mes *message.Message) (err error) {
	up := process.UserProcess{
		Conn: pr.Conn,
	}

	switch mes.Type {
	case message.LoginMesType:
		err = up.ServerProcessLogin(mes)
		if err != nil {
			fmt.Println("ServerProcessLogin err=", err)
			return
		}
	case message.RegisterMesType:
		//处理注册
	default:
		fmt.Println("消息类型不存在，无法处理...")
	}
	return
}

// 处理与客户端的通讯
func (pr *Processor) processTCP() (err error) {
	defer pr.Conn.Close()

	tf := utils.Transfer{
		Conn: pr.Conn,
	}

	for {
		mes, err := tf.ReadPkg()
		if err != nil {
			if err == io.EOF {
				fmt.Println("客户端退出，服务器端也退出")
				return err
			} else {
				fmt.Println("readPkg err", err)
				return err
			}
		}

		err = pr.ServeProcessMes(&mes)
		if err != nil {
			fmt.Println("serverProcessMes err=", err)
			return err
		}
	}

}
