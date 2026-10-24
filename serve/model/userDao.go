package model

import (
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

// 在服务器启动后，就初始化一个UserDao实例
// 做成全局变量，在需要和redis操作时，可直接使用
var (
	MyUserDao *UserDao
)

type UserDao struct {
	pool *redis.Pool
}

// 使用工厂模式，创建一个UserDao实例
func NewUserDao(pool *redis.Pool) (userDao *UserDao) {
	userDao = &UserDao{
		pool: pool,
	}
	return
}

// 根据用户ID，返回一个User实例+Err
func (ud *UserDao) getUserById(conn redis.Conn, id int) (user *User, err error) {
	//给定Id,去redis中查询该用户
	res, err := redis.String(conn.Do("hget", "users", id))
	if err != nil {
		//表示在users哈希中，没有找到对应Id
		if err == redis.ErrNil {
			err = ERROR_USER_NOTEXISTS
			return user, ERROR_USER_NOTEXISTS
		}
		return user, err
	}
	user = &User{}
	//把res反序列化成User实例
	err = json.Unmarshal([]byte(res), user)
	if err != nil {
		fmt.Println("json unmardhal err=", err)
		return
	}
	return user, nil
}

// 完成登录的校验
func (ud *UserDao) Login(userId int, userPwd string) (user *User, err error) {
	//从链接池中取一个连接
	conn := ud.pool.Get()
	defer conn.Close()

	user, err = ud.getUserById(conn, userId)
	if err != nil {
		fmt.Println("getUserById err=", err)
		return
	}

	//这时证明这个用户是获取到了，下一步判断密码
	if user.UserPwd != userPwd {
		err = ERROR_USER_PWD
		return
	}

	return

}
