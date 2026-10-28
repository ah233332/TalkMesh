package main

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

var pool *redis.Pool

// 通过 address、maxIdle、maxActive 和 idleTimeout 构造 redigo 连接池，并赋值给全局 pool。
// Dial 回调使用 redis.Dial("tcp", address) 创建连接，后续 UserDao 可通过该连接池复用 Redis 连接。
func initPool(address string, maxIdle, maxActive int, idleTimeout time.Duration) {
	pool = &redis.Pool{
		MaxIdle:   maxIdle,   //最大空闲链接数
		MaxActive: maxActive, //表示和数据库的最大链接数，0为不限制
		//操作系统依旧会限制，但代码不限制
		IdleTimeout: idleTimeout, //最大空闲时间

		//初始化链接的代码
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", address)
		},
	}
}
