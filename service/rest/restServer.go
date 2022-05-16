package rest

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/p1nant0m/xdp-tracing/service"
)

const (
	REDIS_GET_COMM      = iota
	REDIS_QUERY_TIMEOUT = time.Second * 3
)

func RestServe(ctx context.Context) {

	redisService := ctx.Value("redis-service").(*service.RedisService)
	RedisCommGetHandler := prepareRedisGetHandler(redisService)
	getAllSessionHandler := preparegetAllSessionHandler(redisService)
	r := gin.Default()
	r.GET("test/redis/get/:key", RedisCommGetHandler)
	r.GET("get/all/session", getAllSessionHandler)
	go r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	select {
	case <-ctx.Done():
		os.Exit(0)
	default:
	}
}

type lookup struct {
	Key *service.Key
	ID  string
}

func preparegetAllSessionHandler(redisService *service.RedisService) (fn gin.HandlerFunc) {
	fn = func(c *gin.Context) {
		ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(REDIS_QUERY_TIMEOUT))
		defer cancel()

		uuID := uuid.New().String()
		redisService.Register(uuID)
		defer redisService.Destory(uuID)
		notifyCh, _ := redisService.RetrieveChannel(uuID)
		task := func(rdb *redis.Client) (interface{}, error) {
			value, err := rdb.SMembers(ctx, "sessions").Result()
			return value, err
		}
		ResultType := "[]string"
		redisService.TaskAssign(task, ResultType, uuID)

		select {
		case notifyMsg := <-notifyCh:
			if notifyMsg.ErrorMsg == redis.Nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "capturer has not captured any packet yet",
				})
			} else {
				var key_list []*lookup
				if notifyMsg.ResultType != "[]string" {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": fmt.Sprintf("inconsitency between expeted Type []string and received Type %v", notifyMsg.ResultType),
					})
				}
				sessionList := notifyMsg.ExecuteResult.([]string)
				for _, session := range sessionList {
					key := service.DecodeKey(session)
					key_list = append(key_list, &lookup{
						Key: key,
						ID:  base64.StdEncoding.EncodeToString([]byte(session)),
					})
				}
				c.JSON(http.StatusOK, gin.H{"sessions": key_list})
			}
		case <-ctx.Done():
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "timeout happens when quering redisDB",
			})
		}
	}
	return
}

// template handler for redis querying
func prepareRedisGetHandler(redisService *service.RedisService) (fn gin.HandlerFunc) {
	fn = func(c *gin.Context) {
		ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(REDIS_QUERY_TIMEOUT))
		defer cancel()

		key := c.Param("key")
		uuID := uuid.New().String()
		redisService.Register(uuID)
		defer redisService.Destory(uuID)
		notifyCh, _ := redisService.RetrieveChannel(uuID)
		task := func(rdb *redis.Client) (interface{}, error) {
			value, err := rdb.Get(ctx, key).Result()
			return value, err
		}
		ResultType := "string"
		redisService.TaskAssign(task, ResultType, uuID)

		select {
		case notifyMsg := <-notifyCh:
			if notifyMsg.ErrorMsg == redis.Nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("no related value to the key %v", key),
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"Duration":      notifyMsg.Duration,
					"ResultType":    notifyMsg.ResultType,
					"Client":        notifyMsg.Client,
					"ExecuteResult": notifyMsg.ExecuteResult.(string),
				})
			}
		case <-ctx.Done():
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "timeout happens when quering redisDB",
			})
		}

	}
	return
}
