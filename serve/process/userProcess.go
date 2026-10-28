package process

//处理有关客户的信息功能

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
	"github.com/ah233332/TalkMesh/serve/model"
)

type UserProcess struct {
	Conn net.Conn
}

// 通过 json.Unmarshal 将 mes.Data 还原为 LoginMes，取出 UserId 和 UserPwd，并调用 MyUserDao.Login 查询 Redis。
// 根据返回的 user 和 err 填充 LoginResMes.Code 与 Error，区分登录成功、密码错误、用户不存在和未知错误。
// 最后将 LoginResMes 序列化到外层 Message.Data，并通过 Transfer.WritePkg 把处理结果返回客户端。
func (u *UserProcess) ServerProcessLogin(mes *message.Message) (err error) {
	//1.先从mes中取出mes data,并直接反序列化成LoginMes
	var loginmes message.LoginMes
	err = json.Unmarshal([]byte(mes.Data), &loginmes)
	if err != nil {
		fmt.Println("unmarshal mesData err=", err)
		return
	}

	//声明返回信息,并赋值
	var loginResMes message.Message
	loginResMes.Type = message.LoginResMesType

	var loginResMesData message.LoginResMes

	//需要到redis数据库去验证登录
	user, err := model.MyUserDao.Login(loginmes.UserId, loginmes.UserPwd)

	if err != nil {
		switch err {
		case model.ERROR_USER_PWD:
			loginResMesData.Code = 403 //密码错误
			loginResMesData.Error = err.Error()
		case model.ERROR_USER_NOTEXISTS:
			loginResMesData.Code = 500 //用户不存在
			loginResMesData.Error = err.Error()
		default:
			loginResMesData.Code = 505 //未知错误
			loginResMesData.Error = "未定义的错误"
		}
	} else {
		loginResMesData.Code = 200 //登录成功
		fmt.Println(*user, "登录成功！！！！！！！！")
	}

	data, err := json.Marshal(loginResMesData)
	if err != nil {
		fmt.Println("marshal loginResMesData err=", err)
		return
	}
	//data为序列化之后的loginresmesdata

	loginResMes.Data = string(data)
	//loginResMes赋值完成，准备进行序列化发送

	//返回信息序列化
	data, err = json.Marshal(loginResMes)
	if err != nil {
		fmt.Println("marshal loginResMes err=", err)
		return
	}
	//此时data就是要发送出的内容

	tf := &utils.Transfer{
		Conn: u.Conn,
	}
	//发送data返回到客户端
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("writePkg err=", err)
		return
	}
	return

}

func (u *UserProcess) ServerProcessRegister(mes *message.Message) (err error) {
	//1.把mes数据反序列化整出来
	var registerMes message.RegisterMes
	err = json.Unmarshal([]byte(mes.Data), &registerMes)
	if err != nil {
		fmt.Println("unmarshal mes data err", err)
		return
	}

	//2.设置好ResMes信息
	var registerResMes message.Message
	registerResMes.Type = message.RegisterResMesType

	var registerResMesData message.RegisterResMes

	//3.进行函数操作
	if registerMes.User.UserId <= 0 || registerMes.User.UserPwd == "" || registerMes.User.UserName == "" {
		registerResMesData.Code = 401
		registerResMesData.Error = "用户ID必须为正数，密码和用户名不能为空"
		// 这里不要 return，下面还要把 401 发回客户端
	} else {
		err = model.MyUserDao.Register(&registerMes.User)

		switch err {
		case nil:
			registerResMesData.Code = 200 //表示注册成功
			fmt.Println("注册成功！！！")
			fmt.Printf("您的id=%d,Name=%s,Pwd=%s\n", registerMes.User.UserId, registerMes.User.UserName, registerMes.User.UserPwd)
		case model.ERROR_USER_EXISTS:
			registerResMesData.Code = 400 //表示注册ID被占用
			registerResMesData.Error = err.Error()
		default:
			registerResMesData.Code = 505 //表示其他错误情况
			registerResMesData.Error = "未定义的错误"
		}
	}

	//4.序列化返回的信息并发送到客户端
	data, err := json.Marshal(registerResMesData)
	if err != nil {
		fmt.Println("marshal registerResMesData err=", err)
		return
	}
	registerResMes.Data = string(data)
	//registerResMes赋值完毕，准备序列化

	data, err = json.Marshal(registerResMes)
	if err != nil {
		fmt.Println("marshal registerResMes err=", err)
		return
	}
	//此时data就是准备发送的内容

	tf := utils.Transfer{
		Conn: u.Conn,
	}

	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("tf write err=", err)
		return
	}
	return
}
