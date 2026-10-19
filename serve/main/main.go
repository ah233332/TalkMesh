package main

//主要功能
//1.监听
//2.等待客户端的连接
//3.初始化的工作

import (
	"fmt"
	"io"
	"net"

	"github.com/ah233332/TalkMesh/common/utils"
	"github.com/ah233332/TalkMesh/serve/process"
)

// 处理与客户端的通讯
func processTCP(conn net.Conn) {
	defer conn.Close()

	for {
		mes, err := utils.ReadPkg(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Println("客户端退出，服务器端也退出")
				return
			} else {
				fmt.Println("readPkg err", err)
				return
			}
		}

		err = process.ServerProcessLogin(conn, &mes)
		if err != nil {
			fmt.Println("serverProcessLogin err=", err)
			return
		}
	}
}

func main() {
	fmt.Println("服务器在8899端口监听....")
	listen, err := net.Listen("tcp", "0.0.0.0:8899")
	if err != nil {
		fmt.Println("net.listen err=", err)
		return
	}
	defer listen.Close()

	for {
		fmt.Println("等待连接中...")
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listen accept err=", err)
		}

		go processTCP(conn)
	}
}
