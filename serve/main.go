package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

// 处理与客户端的通讯
func ProcessTCP(conn net.Conn) {
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

		err = serverProcessLogin(conn, &mes)
		if err != nil {
			fmt.Println("serverProcessLogin err=", err)
			return
		}
	}
}

// 登录功能的核心代码
func serverProcessLogin(conn net.Conn, mes *message.Message) (err error) {
	//1.先从mes中取出mes data,并直接反序列化成LoginMes
	var loginmes message.LoginMes
	err = json.Unmarshal([]byte(mes.Data), &loginmes)
	if err != nil {
		fmt.Println("unmarshal mesData err=", err)
		return
	}

	//声明返回信息,并赋值
	var loginResMes message.Message
	loginResMes.Type = message.LoginResMesType

	var loginResMesData message.LoginResMes

	//测试：如果用户ID为114514 密码qwe 认为合法
	if loginmes.UserId == 114514 && loginmes.UserPwd == "qwe" {
		loginResMesData.Code = 200 //登录成功
	} else {
		loginResMesData.Code = 500 //登录失败，用户不存在
		loginResMesData.Error = "该用户不存在，请注册再登录"
	}

	data, err := json.Marshal(loginResMesData)
	if err != nil {
		fmt.Println("marshal loginResMesData err=", err)
		return
	}
	//data为序列化之后的loginresmesdata

	loginResMes.Data = string(data)
	//loginResMes赋值完成，准备进行序列化发送

	//返回信息序列化
	data, err = json.Marshal(loginResMes)
	if err != nil {
		fmt.Println("marshal loginResMes err=", err)
		return
	}
	//此时data就是要发送出的内容

	//发送data返回到客户端
	err = utils.WritePkg(conn, data)
	if err != nil {
		fmt.Println("writePkg err=", err)
		return
	}
	return

}

// 路由功能,根据客户端发送的消息种类不同，决定调用哪个函数处理
func ServeProcessMes(conn net.Conn, mes *message.Message) (err error) {
	switch mes.Type {
	case message.LoginMesType:
		serverProcessLogin(conn, mes)
	case message.RegisterMesType:
		//处理注册
	default:
		fmt.Println("消息类型不存在，无法处理...")
	}
	return
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

		go ProcessTCP(conn)
	}
}
