package model

import (
	"net"

	"github.com/ah233332/TalkMesh/common/message"
)

// 客户端用来维护Conn和用户信息的
type CurUser struct {
	message.User
	Conn net.Conn
}
