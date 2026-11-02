package process

import (
	"fmt"
	"sync"

	"github.com/ah233332/TalkMesh/client/model"
	"github.com/ah233332/TalkMesh/common/message"
)

// 客户端需要维护的东西
var userMgr *UserMgr
var curUser model.CurUser

type UserMgr struct {
	onlineUsers map[int]*message.User
	mu          sync.RWMutex
}

func init() {
	userMgr = &UserMgr{
		onlineUsers: make(map[int]*message.User, 10),
	}
}

// 拷贝在线用户列表副本
func (um *UserMgr) Snapshot() []*message.User {
	um.mu.RLock()
	defer um.mu.RUnlock()

	var list []*message.User
	for _, up := range um.onlineUsers {
		list = append(list, up)
	}
	return list
}

// 在客户端显示当前在线用户
func (um *UserMgr) outputOnlineUsers() {
	fmt.Println("当前用户列表：")
	for _, user := range um.Snapshot() {
		fmt.Printf("用户id:%d  状态：%d\n", user.UserId, user.UserStatus)
	}
}

// 处理服务端返回的NotifyUserStatusMes信息，路由作用并打印列表
func (um *UserMgr) updateUserStatus(notifyUserStatusMes *message.NotifyUserStatusMes) {

	switch notifyUserStatusMes.Status {
	case message.UserOnline:
		//调用发放
		um.markOnline(notifyUserStatusMes.UserId)
	case message.UserOffline:
		um.deleteUser(notifyUserStatusMes.UserId)
	default:
		//fmt.Println("未知状态") //"锁里不做IO"
	}

	//输出在线用户列表
	um.outputOnlineUsers()
}

func (um *UserMgr) markOnline(userId int) {
	um.mu.Lock()
	defer um.mu.Unlock()

	user, ok := um.onlineUsers[userId]
	if ok {
		user.UserStatus = message.UserOnline
		return
	}

	um.onlineUsers[userId] = &message.User{
		UserId:     userId,
		UserStatus: message.UserOnline,
	}
}

// 从在线列表中删除目标用户，带锁方法
func (um *UserMgr) deleteUser(userId int) {
	um.mu.Lock()
	defer um.mu.Unlock()

	delete(um.onlineUsers, userId)
}
