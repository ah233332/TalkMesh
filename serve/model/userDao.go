package model

import (
	"encoding/json"
	"fmt"

	"github.com/ah233332/TalkMesh/common/message"
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

// 通过传入的 Redis 连接池 pool 构造 UserDao，并返回其指针。
// 该对象将连接池保存在内部，后续查询用户和校验密码时可以直接复用连接。
func NewUserDao(pool *redis.Pool) (userDao *UserDao) {
	userDao = &UserDao{
		pool: pool,
	}
	return
}

// 通过 conn.Do("hget", "users", id) 从 users 哈希中读取指定 ID 的 JSON 字符串。
// 命中后使用 json.Unmarshal 将其还原为 User；遇到 redis.ErrNil 时转换为 ERROR_USER_NOTEXISTS，其他错误原样返回。
func (ud *UserDao) getUserById(conn redis.Conn, id int) (user *message.User, err error) {
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
	user = &message.User{}
	//把res反序列化成User实例
	err = json.Unmarshal([]byte(res), user)
	if err != nil {
		fmt.Println("json unmardhal err=", err)
		return
	}
	return user, nil
}

// 通过 ud.pool.Get 获取 Redis 连接，并在 defer 中关闭，然后调用 getUserById 查询 userId。
// 查询成功后比较 user.UserPwd 与 userPwd，不一致时返回 ERROR_USER_PWD，一致时返回 user 供上层生成登录响应。
func (ud *UserDao) Login(userId int, userPwd string) (user *message.User, err error) {
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

func (ud *UserDao) Register(user *message.User) (err error) {
	conn := ud.pool.Get()
	defer conn.Close()

	_, err = ud.getUserById(conn, user.UserId)
	//err为空说明数据库找到了相同id用户
	//err为ERROR_USER_NOTEXISTS，说明改id未注册
	switch err {
	case nil:
		return ERROR_USER_EXISTS
		//还返回了找到的已被注册的用户实例
	case ERROR_USER_NOTEXISTS:
		//可以进行注册

		data, err := json.Marshal(user)
		if err != nil {
			fmt.Println("marshal newuser err=", err)
			return err
		}

		_, err = conn.Do("hset", "users", user.UserId, string(data))
		if err != nil {
			fmt.Println("conn do hset err=", err)
			return err
		}
	default:
		return err
	}
	return nil
}
