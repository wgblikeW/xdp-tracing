package service

import (
	"context"

	"github.com/go-redis/redis"
)

func Register() {

}

func Service() {
	ctx := context.Background()

	rdb := redis.NewClient(MakeNewRedisOptions())

}
