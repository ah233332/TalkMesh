package process

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

type UserProcess struct {
}

func (up *UserProcess) Login(userId int, userPwd string) (err error) {
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

	fmt.Println(string(data))

	tf := utils.Transfer{
		Conn: conn,
	}

	//4.发送数据长度与本体给对方
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("login writePkg err=", err)
		return
	}

	//5.处理服务器端返回的消息
	resMes, err := tf.ReadPkg()
	var loginResMes message.LoginResMes
	err = json.Unmarshal([]byte(resMes.Data), &loginResMes)
	if err != nil {
		fmt.Println("unmarshal err=", err)
		return
	}

	if loginResMes.Code == 200 {
		fmt.Println("登录成功！！！")

		go continueProcess(conn)

		for {
			ShowMenu()
		}
	} else {
		fmt.Println(loginResMes.Error)
	}

	return
}
