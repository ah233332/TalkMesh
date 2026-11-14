package main

import (
	"fmt"
	"strconv"

	"github.com/ah233332/TalkMesh/client/process"
	"github.com/ah233332/TalkMesh/common/utils"
)

// 通过循环显示控制台主菜单，读取 key 并将用户操作分派到登录、注册或退出流程。
// 登录和注册时会收集 userId、userPwd、userName，调用 UserProcess 的对应方法，并输出返回的 err。
func main() {
	var (
		userId   int
		userPwd  string
		userName string
		loop     bool = false
		up       process.UserProcess
	)
	for {
		fmt.Println("---------------用户通讯系统--------------")
		fmt.Println("---1.登录")
		fmt.Println("---2.注册")
		fmt.Println("---3.退出")
		fmt.Print("请输入你的选择：")
		key, err := strconv.Atoi(utils.Readline())
		if err != nil {
			return
		}

		switch key {
		case 1:
			fmt.Println("请输入用户Id：")
			userId, err = strconv.Atoi(utils.Readline())
			if err != nil {
				fmt.Println("用户ID必须是数字")
				continue
			}

			fmt.Println("请输入密码：")
			userPwd = utils.Readline()
			err = up.Login(userId, userPwd)
			if err != nil {
				fmt.Println("登录失败")
			}
		case 2:
			fmt.Println("请输入用户Id：")
			userId, err = strconv.Atoi(utils.Readline())
			if err != nil {
				fmt.Println("用户ID必须是数字")
				continue
			}
			fmt.Println("请输入密码：")
			userPwd = utils.Readline()
			fmt.Println("请输入用户名：")
			userName = utils.Readline()
			err = up.Register(userId, userPwd, userName)
			if err != nil {
				fmt.Println("登录失败")
			}
		case 3:
			fmt.Println("退出")
			loop = true
		default:
			fmt.Println("输入错误")
		}

		if loop {
			fmt.Println("退出成功！")
			break
		}

	}

}
