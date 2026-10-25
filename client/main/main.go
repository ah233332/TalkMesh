package main

import (
	"fmt"

	"github.com/ah233332/TalkMesh/client/process"
)

func main() {
	var (
		key      int
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
		fmt.Scan(&key)

		switch key {
		case 1:
			fmt.Println("请输入用户Id：")
			fmt.Scan(&userId)
			fmt.Println("请输入密码：")
			fmt.Scan(&userPwd)
			err := up.Login(userId, userPwd)
			if err != nil {
				fmt.Println("登录失败")
			}
		case 2:
			fmt.Println("请输入用户Id：")
			fmt.Scan(&userId)
			fmt.Println("请输入密码：")
			fmt.Scan(&userPwd)
			fmt.Println("请输入用户名：")
			fmt.Scan(&userName)
			err := up.Register(userId, userPwd, userName)
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
