package login

import "fmt"

func Login(userId int, userPwd string) (err error) {
	fmt.Printf("userID=%d userPwd=%v", userId, userPwd)
	return nil
}
