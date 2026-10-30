package message

const (
	LoginMesType            = "LoginMes"
	LoginResMesType         = "LoginResMes"
	RegisterMesType         = "RegisterMes"
	RegisterResMesType      = "RegisterResMes"
	NotifyUserStatusMesType = "NotifyUserStatusMes"
	SmsMessageType          = "SmsMessage"
	SmsTransferMesType      = "SmsTransferMes"
)

type Message struct {
	Type string `json:"type"` //消息类型
	Data string `json:"data"` //消息内容
}

// 登录信息
type LoginMes struct {
	UserId   int    `json:"userId"`   //用户ID
	UserPwd  string `json:"userPwd"`  //用户密码
	UserName string `json:"userName"` //用户名字
}

// 登录返回信息
type LoginResMes struct {
	UserIds []int  `json:"userIds"` //增加字段，保存用户id的切片
	Code    int    `json:"code"`    //返回状态码
	Error   string `json:"error"`   //错误信息
	//500表示用户未注册 200表示登录成功 403密码错误 505未知错误
}

// 注册信息
type RegisterMes struct {
	User User `json:"user"`
}

// 注册返回信息
type RegisterResMes struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
	//400表示用户已存在 401代表申请注册的信息不合规
	//200表示注册成功  505表示未知错误
}

// 定义一些状态常量
const (
	UserOnline = iota
	UserOffline
	UserBusyStatus
)

// 状态变化信息
// 为了配合服务端向客户端推送用户状态变化的消息
type NotifyUserStatusMes struct {
	UserId int `json:"userId"`
	Status int `json:"status"` //用户状态

}

// 群发或私聊的信息
type SmsMessage struct {
	UserId     int    `json:"userId"`
	UserName   string `json:"userName"`
	UserStatus int    `json:"status"`
	Content    string `json:"content"`
}

// 服务端处理群聊信息后，发送给其他人的信息类型
type SmsTransferMes struct {
	UserId     int    `json:"userId"`
	UserName   string `json:"userName"`
	UserStatus int    `json:"status"`
	Content    string `json:"content"`
}
