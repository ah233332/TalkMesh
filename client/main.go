package main

import (
	"fmt"
)

func main() {
	var (
		key     int
		userId  int
		userPwd string
		loop    bool = false
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
			fmt.Println("登录")
		case 2:
			fmt.Println("注册")
		case 3:
			fmt.Println("退出")
			loop = true
		default:
			fmt.Println("输入错误")
		}

		if key == 1 {
			fmt.Println("请输入用户名：")
			fmt.Scan(&userId)
			fmt.Println("请输入密码：")
			fmt.Scan(&userPwd)
			err := login(userId, userPwd)
			if err != nil {
				fmt.Println("登录失败")
			}
		} else if key == 2 {
			fmt.Println("进行用户注册")
		}

		if loop {
			fmt.Println("退出成功！")
			break
		}

	}

}
