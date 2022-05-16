package rest

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/p1nant0m/xdp-tracing/service"
)

const (
	REDIS_GET_COMM = iota
)

const (
	CLIENT_NAME_REST = "rest"
)

func RestServe(ctx context.Context) {

	redisService := ctx.Value("redis-service").(*service.RedisService)
	RedisCommGetHandler := execRedisCommGET(redisService)
	r := gin.Default()
	r.GET("/redis/get/:key", RedisCommGetHandler)

	go r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	select {
	case <-ctx.Done():
		os.Exit(0)
	default:
	}
}

func execRedisCommGET(redisService *service.RedisService) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		key := c.Param("key")
		uuID := uuid.New().String()
		redisService.Register(uuID)
		defer redisService.Destory(uuID)
		notifyCh, _ := redisService.RetrieveChannel(uuID)
		task := func(rdb *redis.Client) (interface{}, error) {
			value, err := rdb.Get(c, key).Result()
			return value, err
		}
		ResultType := "string"
		redisService.TaskAssign(task, ResultType, "RedisGet")
		notifyMsg := <-notifyCh
		if notifyMsg.ErrorMsg != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": notifyMsg.ErrorMsg,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"Duration":      notifyMsg.Duration,
				"ResultType":    notifyMsg.ResultType,
				"Client":        notifyMsg.Client,
				"ExecuteResult": notifyMsg.ExecuteResult.(string),
			})
		}
	}
	return fn
}
