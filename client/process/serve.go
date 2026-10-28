package process

import (
	"fmt"
	"net"
	"os"

	"github.com/ah233332/TalkMesh/common/utils"
)

// 通过控制台打印登录后的功能菜单，读取 key，并按选项进入在线用户、发送消息、消息列表或退出流程。
// 当前除退出外的选项均为占位输出，选择 4 时会调用 os.Exit(0) 结束客户端进程。
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

// 通过 utils.Transfer 持续调用 ReadPkg 读取服务端推送的数据包，读取成功后输出 mes。
// 一旦读取失败，会记录 err 并返回，从而结束当前接收消息的 goroutine。
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
