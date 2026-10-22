package model

import (
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type UserDao struct {
	pool *redis.Pool
}

// UserDao的作用

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
