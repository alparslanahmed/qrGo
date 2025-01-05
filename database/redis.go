package database

import (
	"alparslanahmed/qrGo/config"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Config("REDIS_URL"),
		Password: "",
		DB:       0,
	})
}
