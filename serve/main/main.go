package main

//主要功能
//1.监听
//2.等待客户端的连接
//3.初始化的工作

import (
	"fmt"
	"net"
	"time"

	"github.com/ah233332/TalkMesh/serve/model"
)

func initUserDao() {
	model.MyUserDao = model.NewUserDao(pool)
}

func main() {
	//初始化redis连接池
	initPool("localhost:6379", 16, 0, 300*time.Second)
	//初始化UserDao
	initUserDao()

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
			continue
		}

		pr := Processor{
			Conn: conn,
		}

		go pr.processTCP()
	}
}
