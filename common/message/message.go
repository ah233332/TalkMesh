package message

const (
	LoginMesType    = "LoginMes"
	LoginResMesType = "LoginResMes"
)

type Message struct {
	Type string `json:"type"` //消息类型
	Data string `json:"data"` //消息内容
}

type LoginMes struct {
	UserId   int    `json:"userId"`   //用户ID
	UserPwd  string `json:"userPwd"`  //用户密码
	UserName string `json:"userName"` //用户名字
}

type LoginResMes struct {
	Code  int    `json:"code"`  //返回状态码  500表示用户未注册 200表示登录成功
	Error string `json:"error"` //错误信息
}
