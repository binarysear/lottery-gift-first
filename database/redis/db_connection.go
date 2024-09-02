package redis

import (
	"context"
	"gift/util"
	"github.com/go-redis/redis/v8"
	"sync"
)

var (
	gift_redis      *redis.Client
	gift_redis_once sync.Once
)

// 创建连接
func createRedisClient(address, passwd string, db int) *redis.Client {
	cli := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: passwd,
		DB:       db,
	})
	if err := cli.Ping(context.Background()).Err(); err != nil {
		util.LogRus.Panicf("connect to redis %d failed %v", db, err)
	} else {
		util.LogRus.Infof("connect to redis %d", db)
	}
	return cli
}

// 获取redis连接
func GetRedisClient() *redis.Client {
	gift_redis_once.Do(func() {
		if gift_redis == nil {
			viper := util.CreateConfig("redis")
			addr := viper.GetString("addr")
			pass := viper.GetString("password")
			db := viper.GetInt("db")
			gift_redis = createRedisClient(addr, pass, db)
		}
	})

	return gift_redis
}
