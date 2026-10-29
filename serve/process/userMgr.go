package process

import "fmt"

var (
	userMgr *UserMgr
)

type UserMgr struct {
	onlineUsers map[int]*UserProcess
}

func init() {
	userMgr = &UserMgr{
		onlineUsers: make(map[int]*UserProcess, 1024),
	}
}

// 对onlineUsers添加/修改
func (um *UserMgr) AddOnlineUser(up *UserProcess) {
	um.onlineUsers[up.UserId] = up
}

// 对onlineUsers删除
func (um *UserMgr) DelOnlineUser(userId int) {
	delete(um.onlineUsers, userId)
}

// 根据id返回对应的onlineUsers
func (um *UserMgr) GetOnlineUserById(userid int) (up *UserProcess, err error) {
	up, ok := um.onlineUsers[userid]
	if !ok {
		//找的用户当前不在线
		err = fmt.Errorf("用户%d 不在线", userid)
		return
	}
	return
}

// 返回当前所有在线用户
func (um *UserMgr) GetAllOnlineUser() map[int]*UserProcess {
	return um.onlineUsers
}
