package config

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func InitRedis(ctx context.Context) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		return err
	}

	fmt.Println("Conexión OK:", pong)

	Rdb = rdb
	return nil
}
