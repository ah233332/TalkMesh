package model

import "errors"

//根据业务需要，自定义一些错误
var (
	ERROR_USER_NOTEXISTS = errors.New("用户不存在")
	ERROR_USER_EXISTS    = errors.New("用户已存在")
	ERROR_USER_NOTLOGIN  = errors.New("用户未登录")
	ERROR_USER_PWD       = errors.New("密码错误")
	ERROR_USER_NOTONLINE = errors.New("该用户不在线")
	ERROR_USER_MISMATCH  = errors.New("发送者身份不一致")
	ERROR_SEND_USER      = errors.New("发送对象错误")
	ERROR_SERVE_REDIS    = errors.New("redis服务错误")
)

//自定义一些成功
var (
	SUS_SEND_OFFLINE = errors.New("成功发送离线消息")
)
