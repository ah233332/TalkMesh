package process

import (
	"fmt"
	"sync"
)

var (
	userMgr *UserMgr
)

type UserMgr struct {
	onlineUsers map[int]*UserProcess
	mu          sync.RWMutex //读写互斥锁
	//lock unlock是写锁，只能单独一人写
	//rlock runlock是读锁，可以多人同时读
}

func init() {
	userMgr = &UserMgr{
		onlineUsers: make(map[int]*UserProcess, 1024),
	}
}

// 快照作用，把看在线用户与发送信息分开来执行
// 相当于创建一个在线用户的名单副本，不占用互斥锁，之后让程序慢慢群发信息
func (um *UserMgr) Snapshot() []*UserProcess {
	um.mu.RLock()
	defer um.mu.RUnlock()

	list := make([]*UserProcess, 0, len(um.onlineUsers))
	for _, up := range um.onlineUsers {
		list = append(list, up)
	}

	return list
}

// 对onlineUsers添加/修改
func (um *UserMgr) AddOnlineUser(up *UserProcess) {
	um.mu.Lock()
	defer um.mu.Unlock()

	um.onlineUsers[up.UserId] = up

}

// 对onlineUsers删除
func (um *UserMgr) DelOnlineUser(userId int) {
	um.mu.Lock()
	defer um.mu.Unlock()

	delete(um.onlineUsers, userId)
}

func DelOnlineUser(userId int) {
	userMgr.DelOnlineUser(userId)
}

// 根据id返回对应的onlineUsers
func (um *UserMgr) GetOnlineUserById(userid int) (up *UserProcess, err error) {
	um.mu.RLock()
	defer um.mu.RUnlock()

	up, ok := um.onlineUsers[userid]
	if !ok {
		//找的用户当前不在线
		err = fmt.Errorf("用户%d 不在线", userid)
		return
	}
	return
}
