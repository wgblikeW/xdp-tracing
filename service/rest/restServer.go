package rest

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/p1nant0m/xdp-tracing/service"
)

func RestServe(ctx context.Context) {
	r := gin.Default()
	taskCh := ctx.Value("redis-taskCh").(chan<- *service.AssignTask)

	r.GET("/redis/get/:key", execRedisCommGET)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	select {
	case <-ctx.Done():
		os.Exit(0)
	default:
	}
}

func execRedisCommGET(c *gin.Context) {
	key := c.Param("key")
	task := func(rdb *redis.Client) (interface{}, error) {
		value, err := rdb.Get(c, key).Result()
		return value, err
	}
	ResultType := "string"

}
