package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
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
	defer conn.Close()

	//2.将mes的数据填完整
	var mes message.Message
	mes.Type = message.LoginMesType

	var loginMex message.LoginMes
	loginMex.UserId = userId
	loginMex.UserPwd = userPwd

	//将LoginMex结构体序列化之后再填入mes中
	data, err := json.Marshal(loginMex)
	if err != nil {
		fmt.Println("Marshal err=", err)
		return
	}

	mes.Data = string(data)

	//3.将mes序列化
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("Marshal err=", err)
		return
	}
	//data就是要发送的信息

	//4.发送数据长度与本体给对方
	err = utils.WritePkg(conn, data)
	if err != nil {
		fmt.Println("login writePkg err=", err)
		return
	}

	//5.处理服务器端返回的消息
	resMes, err := utils.ReadPkg(conn)
	var loginResMes message.LoginResMes
	err = json.Unmarshal([]byte(resMes.Data), &loginResMes)
	if err != nil {
		fmt.Println("unmarshal err=", err)
		return
	}

	switch loginResMes.Code {
	case 200:
		fmt.Println("登录成功！！！")
	case 500:
		fmt.Println("用户未注册")

	}

	return
}

/*
	//4.在发送mes之前，先发送data字节长度，以防丢包
	//conn.write()只能发送[]byte类型
	//需要先把data长度信息，转成[]byte类型
	var pkgLen uint32
	pkgLen = uint32(len(data))
	var buf [4]byte
	//bigendian littleendian 一个是大端传输，一个小端传输
	//大端按高位在前，小端是反过来低位在前的
	//大端网络使用，小端CPU内存本地用
	//binary包是实现 数字 与 byte切片 之间的转化
	binary.BigEndian.PutUint32(buf[:4], pkgLen)
	//发送长度
	n, err := conn.Write(buf[:4])
	if n != 4 || err != nil {
		fmt.Println("长度发送失败err=", err)
		return
	}

	//5.发送消息本身
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("conn.write data err=", err)
		return
	}
*/
