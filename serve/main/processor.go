package main

//中转路由功能

import (
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/serve/process"
)

// 路由功能,根据客户端发送的消息种类不同，决定调用哪个函数处理
func ServeProcessMes(conn net.Conn, mes *message.Message) (err error) {
	switch mes.Type {
	case message.LoginMesType:
		process.ServerProcessLogin(conn, mes)
	case message.RegisterMesType:
		//处理注册
	default:
		fmt.Println("消息类型不存在，无法处理...")
	}
	return
}
