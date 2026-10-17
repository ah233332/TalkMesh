package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
)

func login(userId int, userPwd string) (err error) {
	//1.连接到服务器端
	conn, err := net.Dial("tcp", "0.0.0.0:8899")
	if err != nil {
		fmt.Println("net dial err=", err)
		return
	} else {
		fmt.Println("成功连接服务器")
	}

	//2.将mes的数据填完整
	var mes message.Message
	mes.Type = message.LoginMesType

	var loginMex message.LoginMes
	loginMex.UserId = userId
	loginMex.UserPwd = userPwd

	//将LoginMex结构体序列化之后再填入mes中
	jsonData, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("Marshal err=", err)
		return
	}

	mes.Data = string(jsonData)

	//3.将mes序列化
	jsonData, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("Marshal err=", err)
		return
	}
	//jsonData就是要发送的信息

	//4.在发送mes之前，先发送数据字节长度，以防丢包

}
