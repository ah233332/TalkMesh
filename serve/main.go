package main

import (
	"fmt"
	"net"
)

// 处理与客户端的通讯
func ProcessTCP(conn net.Conn) {

}

func main() {
	fmt.Println("服务器在8899端口监听....")
	listen, err := net.Listen("tcp", "0.0.0.0:8899")
	if err != nil {
		fmt.Println("net.listen err=", err)
		return
	}

	for {
		fmt.Println("等待连接中...")
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listen accept err=", err)
		}

		go ProcessTCP(conn)
	}
}
