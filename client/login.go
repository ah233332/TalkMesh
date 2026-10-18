package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"

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

	//6.处理服务器端返回的消息
	time.Sleep(20 * time.Second)

	return
}
