package process

import (
	"fmt"

	"github.com/ah233332/TalkMesh/client/model"
	"github.com/ah233332/TalkMesh/common/message"
)

// 客户端需要维护的东西
var onlineUsers map[int]*message.User = make(map[int]*message.User, 10)
var curUser model.CurUser

// 在客户端显示当前在线用户
func outputOnlineUsers() {
	fmt.Println("当前用户列表：")
	for id := range onlineUsers {
		fmt.Println("用户id:\t", id)
	}
}

// 处理服务端返回的NotifyUserStatusMes信息
func updateUserStatus(notifyUserStatusMEs *message.NotifyUserStatusMes) {
	user, ok := onlineUsers[notifyUserStatusMEs.UserId]
	if !ok { //onlineUsers内没有改id时
		user = &message.User{
			UserId: notifyUserStatusMEs.UserId,
		}
	}

	user.UserStatus = notifyUserStatusMEs.Status

	onlineUsers[notifyUserStatusMEs.UserId] = user

	//输出在线用户列表
	outputOnlineUsers()

}
