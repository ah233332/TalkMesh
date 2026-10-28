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

// 通过 net.Dial 建立到 0.0.0.0:8899 的 TCP 连接，将 userId 和 userPwd 写入 LoginMes 并序列化到 Message.Data。
// 随后使用 Transfer.WritePkg 发送登录请求，再用 Transfer.ReadPkg 解析 LoginResMes。
// 登录成功时输出提示、启动 continueProcess 接收消息并循环显示菜单；失败时输出服务端返回的错误信息。
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

	//fmt.Println(string(data))

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

// 通过 net.Dial 建立 TCP 连接，将 userId、userPwd、userName 填入 RegisterMes 并序列化到 Message.Data。
// 发送注册包后使用 Transfer.ReadPkg 获取服务端响应，再按 registerResMes.Code 输出成功、用户已存在或错误信息。
// 任一连接、序列化、发送或响应解析步骤失败时都会返回对应的 err。
func (up *UserProcess) Register(userId int, userPwd string, userName string) (err error) {
	//0.先判断用户要注册的信息是否合法
	if userId <= 0 || userName == "" || userPwd == "" {
		fmt.Println("注册信息不合法，请重新输入")
		return
	}

	//1.连接到服务器端
	conn, err := net.Dial("tcp", "127.0.0.1:8899")
	if err != nil {
		fmt.Println("net dial err=", err)
		return
	} else {
		fmt.Println("成功连接服务器")
	}
	defer conn.Close()

	//2.将mes的数据填完整
	var mes message.Message
	mes.Type = message.RegisterMesType

	var registerMex message.RegisterMes
	registerMex.User.UserId = userId
	registerMex.User.UserPwd = userPwd
	registerMex.User.UserName = userName

	//将registerMex结构体序列化之后再填入mes中
	data, err := json.Marshal(registerMex)
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

	tf := utils.Transfer{
		Conn: conn,
	}

	//4.发送数据长度与本体给对方
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("register writePkg err=", err)
		return
	}

	//5.处理服务器端返回的消息
	registerResMes, err := tf.ReadPkg()
	if err != nil {
		fmt.Println("register readPkg err=", err)
		return
	}

	var registerResMesData message.RegisterResMes
	err = json.Unmarshal([]byte(registerResMes.Data), &registerResMesData)
	if err != nil {
		fmt.Println("unmarshal err=", err)
		return
	}

	switch registerResMesData.Code {
	case 200:
		fmt.Println("注册成功！！！")
	case 400:
		fmt.Println("用户ID已被占用")
	case 401:
		fmt.Println("申请注册的信息非法")
	case 505:
		fmt.Println("未知错误")
	default:
		fmt.Println(registerResMesData.Error)
	}

	return
}
