package process

import (
	"fmt"
	"net"
	"os"

	"github.com/ah233332/TalkMesh/common/utils"
)

// 显示成功登录后的菜单
func ShowMenu() {
	fmt.Println("---------恭喜XXX登录成功---------")
	fmt.Println("-------1.显示在线用户列表--------")
	fmt.Println("-----------2.发送消息-----------")
	fmt.Println("-----------3.消息列表-----------")
	fmt.Println("-----------4.退出系统-----------")
	fmt.Println("请选择(1-4):")

	var key int
	fmt.Scan(&key)
	switch key {
	case 1:
		fmt.Println("111")
	case 2:
		fmt.Println("2222")
	case 3:
		fmt.Println("3333")
	case 4:
		fmt.Println("444")
		os.Exit(0)
	default:
		fmt.Println("输入错误！！！")
	}

}

// 和服务器保持通讯
func continueProcess(Conn net.Conn) {
	tf := utils.Transfer{
		Conn: Conn,
	}

	for {
		fmt.Println("客户端正在等待读取服务器发送的消息")
		mes, err := tf.ReadPkg()
		if err != nil {
			fmt.Println("continue readPkg err=", err)
			return
		}

		//如果读取到消息...
		fmt.Println("mes=", mes)
	}
}
