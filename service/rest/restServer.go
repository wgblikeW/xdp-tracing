package rest

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
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
	ENABLE_PRODUCTION   = true
)

var localIPv4 string

func RestServe(ctx context.Context) {
	if ENABLE_PRODUCTION {
		gin.SetMode(gin.ReleaseMode)
	}
	localIPv4 = localIPObtain()
	redisService := ctx.Value("redis-service").(*service.RedisService)
	RedisCommGetHandler := prepareRedisGetHandler(redisService)
	getAllSessionHandler := preparegetAllSessionHandler(redisService)
	getSessionPackets := preparegetSessionPackets(redisService)

	r := gin.Default()
	r.GET("test/redis/get/:key", RedisCommGetHandler)
	r.GET("get/all/session", getAllSessionHandler)
	r.GET("get/session/:key", getSessionPackets)
	go r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	select {
	case <-ctx.Done():
		os.Exit(0)
	default:
	}
}

func preparegetSessionPackets(redisService *service.RedisService) (fn gin.HandlerFunc) {
	fn = func(c *gin.Context) {
		ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(REDIS_QUERY_TIMEOUT))
		defer cancel()

		key_base64 := c.Param("key")

		uuID := uuid.New().String()
		redisService.Register(uuID)
		defer redisService.Destory(uuID)
		notifyCh, _ := redisService.RetrieveChannel(uuID)

		key, err := base64.URLEncoding.DecodeString(key_base64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "the input is not a valid base64 encoding string",
			})
			return
		}

		task := func(rdb *redis.Client) (interface{}, error) {
			value, err := rdb.ZRangeArgsWithScores(ctx, redis.ZRangeArgs{
				Key:   string(key),
				Start: 0,
				Stop:  -1,
			}).Result()
			return value, err
		}
		ResultType := "[]redis.Z"
		redisService.TaskAssign(task, ResultType, uuID)

		select {
		case notifyMsg := <-notifyCh:
			if notifyMsg.ErrorMsg == redis.Nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "capturer has not captured any packet yet",
				})
			} else {
				var value_list []interface{}
				if notifyMsg.ResultType != "[]redis.Z" {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": fmt.Sprintf("inconsitency between expeted Type []redis.Z and received Type %v", notifyMsg.ResultType),
					})
				}

				ZList := notifyMsg.ExecuteResult.([]redis.Z)
				for _, session := range ZList {
					if packet, err := service.DecodeValue(session.Member.(string)); err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					} else {
						if !packet.PayloadExist {
							packet.Payload = nil
						}
						value_list = append(value_list, struct {
							*service.Value
							Timestamp float64
						}{packet, session.Score})
					}
				}
				c.JSON(http.StatusOK, gin.H{"packets": value_list})
			}
		case <-ctx.Done():
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "timeout happens when quering redisDB",
			})
		}
	}
	return
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
						ID:  base64.URLEncoding.EncodeToString([]byte(session)),
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

func localIPObtain() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
