package process

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
	"github.com/ah233332/TalkMesh/serve/model"
)

type SmsProcess struct {
	Conn     net.Conn
	UserId   int
	UserName string
}

// 处理群聊消息的核心函数方法
func (sp *SmsProcess) ServeGroupMes(mes *message.Message) (err error) {
	//1.先把消息解出来
	var smsMes message.SmsMessage
	err = json.Unmarshal([]byte(mes.Data), &smsMes)
	if err != nil {
		fmt.Println("unmarshal mes data err=", err)
		return
	}

	//安全校验
	if sp.UserId == 0 {
		err = model.ERROR_USER_NOTLOGIN
		return
	}
	smsMes.UserId = sp.UserId
	smsMes.UserName = sp.UserName

	mes.Type = message.SmsTransferMesType
	data, err := json.Marshal(smsMes)
	if err != nil {
		fmt.Println("marshal Smsmes err=", err)
		return
	}
	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal mes err=", err)
		return
	}
	for _, up := range userMgr.Snapshot() {

		errSend := sp.sendMes(up.Conn, data)
		if errSend != nil {
			fmt.Printf("发送给用户:%d 发送失败\n", up.UserId)
			continue
		}
	}

	return
}

// 主要功能用来给在线用户发送信息
func (sp *SmsProcess) sendMes(conn net.Conn, data []byte) (err error) {

	tf := utils.Transfer{
		Conn: conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("sendGroupMes writePkg data err=", err)
		return
	}
	return
}

// 服务私聊的核心代码，转发信息+返回信息
func (sp *SmsProcess) ServeToUserMes(mes *message.Message) (err error) {
	var toUserMes message.SmsToUserMes
	err = json.Unmarshal([]byte(mes.Data), &toUserMes)
	if err != nil {
		fmt.Println("unmarshal mes data err=", err)
		return
	}

	//安全校验
	if sp.UserId == 0 { //是否登录
		err = model.ERROR_USER_NOTLOGIN
		err = sp.toUserResMes(err, mes)
		if err != nil {
			fmt.Println("toUserResMes err=", err)
			return
		}
		return
	}
	if sp.UserId != toUserMes.UserId { //发送者信息是否一致
		err = model.ERROR_USER_MISMATCH
		err = sp.toUserResMes(err, mes)
		if err != nil {
			fmt.Println("toUserResMes err=", err)
			return
		}
		return
	}
	if toUserMes.ToUserId == sp.UserId { //判断发送者是否为自己
		err = model.ERROR_SEND_USER
		err = sp.toUserResMes(err, mes)
		if err != nil {
			fmt.Println("toUserResMes err=", err)
			return
		}
		return
	}

	// 查redis是否存在目标用户
	redisConn := model.MyUserDao.Pool.Get()
	defer redisConn.Close()

	_, redisErr := model.MyUserDao.GetUserById(redisConn, toUserMes.ToUserId)
	if redisErr != nil {
		if redisErr != model.ERROR_USER_NOTEXISTS {
			err = model.ERROR_SERVE_REDIS
			err = sp.toUserResMes(err, mes)
			if err != nil {
				fmt.Println("toUserResMes err=", err)
				return
			}
			return
		} else {
			err = redisErr
			err = sp.toUserResMes(err, mes)
			if err != nil {
				fmt.Println("toUserResMes err=", err)
				return
			}
			return
		}
	}

	//查询目标用户是否在线
	up, err := userMgr.GetOnlineUserById(toUserMes.ToUserId)
	toUserMes.UserName = sp.UserName
	if err != nil {
		//目标用户不在线的情况
		//调用离线函数
		err = sp.transferOffMes(&toUserMes)
		if err != nil {
			//发送离线消息失败
			err = sp.toUserResMes(err, mes)
			if err != nil {
				fmt.Println("toUserResMes err=", err)
				return
			}
			return
		} else {
			//成功发送离线消息
			err = sp.toUserResMes(model.SUS_SEND_OFFLINE, mes)
			if err != nil {
				fmt.Println("toUserResMes err=", err)
				return
			}
			return
		}

	} else {
		//目标用户在线的情况下
		//调用转发函数
		err = sp.transferToUserMes(&toUserMes, up)
		if err != nil {
			//转发信息失败
			err = sp.toUserResMes(err, mes)
			if err != nil {
				fmt.Println("toUserResMes err=", err)
				return
			}
			return
		} else {
			//成功转发信息
			err = sp.toUserResMes(err, mes)
			if err != nil {
				fmt.Println("toUserResMes err=", err)
				return
			}
			return
		}
	}

}

// 服务端转发私信的函数方法
func (sp *SmsProcess) transferToUserMes(toUserMes *message.SmsToUserMes, up *UserProcess) (err error) {

	var mes message.Message
	data, err := json.Marshal(toUserMes)
	if err != nil {
		fmt.Println("marshal toUsermes err=", err)
		return err
	}
	mes.Type = message.SmsToUserTransferMesType
	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal mes err=", err)
		return err
	}

	tf := utils.Transfer{
		Conn: up.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("sendGroupMes writePkg data err=", err)
		return
	}
	return
}

// 服务端返回客户端的响应信息
func (sp *SmsProcess) toUserResMes(err error, mes *message.Message) error {
	//封装返回客户端的信息
	mes.Type = message.SmsToUserResMesType
	var toUserResMes message.SmsToUserResMes
	if err != nil {
		switch err {
		case model.ERROR_USER_NOTONLINE:
			toUserResMes.Code = 400 //目标用户不在线
			toUserResMes.Error = err.Error()
		case model.SUS_SEND_OFFLINE:
			toUserResMes.Code = 201 //成功发送离线消息
			toUserResMes.Error = err.Error()
		case model.ERROR_USER_NOTLOGIN:
			toUserResMes.Code = 401 //用户未登录
			toUserResMes.Error = err.Error()
		case model.ERROR_SEND_USER:
			toUserResMes.Code = 402 //发送对象错误
			toUserResMes.Error = err.Error()
		case model.ERROR_USER_MISMATCH:
			toUserResMes.Code = 403 //发送者身份不一致
			toUserResMes.Error = err.Error()
		case model.ERROR_USER_NOTEXISTS:
			toUserResMes.Code = 405 //目标用户不存在
			toUserResMes.Error = err.Error()

		default:
			toUserResMes.Code = 505 //未知错误
			toUserResMes.Error = err.Error()
		}
	} else {
		toUserResMes.Code = 200 //发送成功
	}

	data, err := json.Marshal(toUserResMes)
	if err != nil {
		fmt.Println("marshal toUserResMes err=", err)
		return err
	}
	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal mes err=", err)
		return err
	}
	//此时data为可直接返回的信息

	tf := utils.Transfer{
		Conn: sp.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("writePkg mes err=", err)
		return err
	}
	return err
}

// 离线消息处理方法
func (sp *SmsProcess) transferOffMes(toUserMes *message.SmsToUserMes) (err error) {
	// !这里可以加入离线消息功能
	//默认已登录 发送对象不是自己 发送者身份一致 目标用户不在线 目标用户存在redis

	//1.从连接池取一根链接
	redisConn := model.MyUserDao.Pool.Get()
	defer redisConn.Close()

	//2.存入redis list
	var mes message.Message
	mes.Type = message.SmsToUserTransferMesType

	data, err := json.Marshal(toUserMes)
	if err != nil {
		fmt.Println("marshal toUserMes err=", err)
		return
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal Mes err=", err)
		return
	}

	key := fmt.Sprintf("offlines:%d", toUserMes.ToUserId)
	value := string(data)

	_, err = redisConn.Do("rpush", key, value)
	if err != nil {
		fmt.Println("redis rpush err=", err)
		return
	}

	//加上限，最多100条离线信息
	_, err = redisConn.Do("ltrim", key, -100, -1)
	if err != nil {
		fmt.Println("redis ltrim err=", err)
		return
	}

	return
}
