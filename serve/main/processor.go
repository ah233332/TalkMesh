package main

//中转路由功能

import (
	"fmt"
	"io"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
	"github.com/ah233332/TalkMesh/serve/process"
)

type Processor struct {
	Conn net.Conn
}

// 通过 mes.Type 判断业务类型，创建持有当前 Conn 的 UserProcess，并将登录消息交给 ServerProcessLogin 处理。
// 注册分支目前只保留接入位置，未知类型会输出提示；各业务处理返回的 err 会继续交给调用方处理。
func (pr *Processor) ServeProcessMes(mes *message.Message) (err error) {

	switch mes.Type {
	case message.LoginMesType:
		up := process.UserProcess{
			Conn: pr.Conn,
		}
		err = up.ServerProcessLogin(mes)
		if err != nil {
			fmt.Println("ServerProcessLogin err=", err)
			return
		}
	case message.RegisterMesType:
		up := process.UserProcess{
			Conn: pr.Conn,
		}
		err = up.ServerProcessRegister(mes)
		if err != nil {
			fmt.Println("ServerProcessRegister err=", err)
			return
		}
	case message.SmsMessageType:
		sp := process.SmsProcess{}
		err = sp.ServeGroupMes(mes)
		if err != nil {
			fmt.Println("ServerGroupMes err=", err)
			return
		}
	default:
		fmt.Println("消息类型不存在，无法处理...")
	}
	return
}

// 通过 utils.Transfer 循环读取客户端数据包，把解析得到的 mes 交给 ServeProcessMes 路由处理。
// 遇到 io.EOF 表示客户端退出，其他读取或处理错误也会记录并返回，同时 defer 关闭当前 Conn。
func (pr *Processor) processTCP() (err error) {
	defer pr.Conn.Close()

	tf := utils.Transfer{
		Conn: pr.Conn,
	}

	for {
		mes, err := tf.ReadPkg()
		if err != nil {
			if err == io.EOF {
				fmt.Println("客户端退出，服务器端也退出")
				return err
			} else {
				fmt.Println("readPkg err", err)
				return err
			}
		}

		err = pr.ServeProcessMes(&mes)
		if err != nil {
			fmt.Println("serverProcessMes err=", err)
			return err
		}
	}

}
