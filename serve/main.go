package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
)

// readPkg 从连接中读取一个完整的数据包。
// 它先读取 4 字节大端长度头，再按该长度读取消息体，
// 最后将 JSON 格式的消息体反序列化为 message.Message。
func readPkg(conn net.Conn) (mes message.Message, err error) {
	buf := make([]byte, 8096)
	fmt.Println("读取客户端发送的数据...")
	_, err = conn.Read(buf[:4])
	if err != nil {
		fmt.Println("conn read err=", err)
		return
	}

	//把buf[:4]转成一个uint32类型
	var pkgLen uint32
	pkgLen = binary.BigEndian.Uint32(buf[:4])

	//根据pkgLen读取消息内容
	n, err := conn.Read(buf[:pkgLen]) //从conn读取pkgLen个字节到buf里面
	if n != int(pkgLen) || err != nil {
		fmt.Println("conn read err=", err)
		return
	}

	//buf  反序列化成 Message类型
	err = json.Unmarshal(buf[:pkgLen], &mes)
	if err != nil {
		fmt.Println("unmarshal err=", err)
		return
	}

	return

}

// 将mes Data 数据反序列化给lmes
func readData(mes message.Message) (lmes message.LoginMes, err error) {
	err = json.Unmarshal([]byte(mes.Data), &lmes)
	if err != nil {
		fmt.Println("unmarshal mesData err=", err)
		return
	}
	return
}

// 处理与客户端的通讯
func ProcessTCP(conn net.Conn) {
	defer conn.Close()

	for {

		mes, err := readPkg(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Println("客户端退出，服务器端也退出")
				return
			}
			fmt.Println("readPkg err", err)
			return
		}

		lmes, err := readData(mes)
		if err != nil {
			fmt.Println("readData err", err)
			return
		}

		fmt.Printf("%v\n%v\n", mes, lmes)
	}
}

func serverProcessLogin(conn net.Conn, mes *message.Message) (err error) {
	//登录功能的核心代码
	//1.先从mes中取出mes data,并直接反序列化成LoginMes
	return
}

// 根据客户端发送的消息种类不同，决定调用哪个函数处理
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
