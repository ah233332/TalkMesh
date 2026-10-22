package model

type User struct {
	//用户信息的json字符串的key 和 结构体的字段对应的tag名 需一致！！！
	UserId   int    `json:"userId"`
	UserPwd  string `json:"userPwd"`
	UserName string `json:"userName"`
}
