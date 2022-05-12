package service

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func Service() {
	ctx := context.Background()

	rdb := redis.NewClient(MakeNewRedisOptions())

	err := rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := rdb.Get(ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := rdb.Get(ctx, "key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}

	result, _ := rdb.HSet(ctx, "tcp_info", []string{"SrcIP", "192.168.176.1", "SrcPort", "1080"}).Result()
	fmt.Println(result)
	val3, _ := rdb.HGetAll(ctx, "tcp_info").Result()
	fmt.Println(val3)
}
