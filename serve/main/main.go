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

// 通过全局 pool 调用 NewUserDao 创建用户数据访问对象，并赋值给 model.MyUserDao。
// 后续登录等业务可以直接复用该对象访问 Redis，无需重复创建连接池。
func initUserDao() {
	model.MyUserDao = model.NewUserDao(pool)
}

// 通过 initPool 和 initUserDao 初始化 Redis 相关依赖，再使用 net.Listen 监听 0.0.0.0:8899。
// 循环调用 Accept 接收 conn，为每个连接创建 Processor 并启动 goroutine 执行 processTCP，从而并发处理多个客户端。
func main() {
	//初始化redis连接池
	initPool("localhost:6379", 16, 0, 300*time.Second)
	//初始化UserDao
	initUserDao()

	fmt.Println("服务器在8899端口监听....")
	listen, err := net.Listen("tcp", "127.0.0.1:8899")
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
