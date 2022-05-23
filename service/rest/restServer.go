package rest

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/p1nant0m/xdp-tracing/handler/utils"
	"github.com/p1nant0m/xdp-tracing/service"
)

const (
	REDIS_GET_COMM      = iota
	REDIS_QUERY_TIMEOUT = time.Second * 3
	REDIS_SERVICE       = "redis-service"
)

var localIPv4 string
var ENABLE_PRODUCTION bool
var restConfig *service.RestConfig

func RunRestServer(configPath string) {
	// Setup notifier and Make Configuration of All Services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watcher := make(chan os.Signal, 1)
	signal.Notify(watcher, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-watcher
		// OS Signal Catched, exit the program gracefully
		cancel()
	}()

	err := service.ReadAndParseConfig(configPath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	restConfig = service.ExtractRestConfig()

	// Setup Redis Service
	var redisService service.Service = service.NewRedisService(ctx)
	redisService.Conn()
	go redisService.Serve()

	// Start Rest Server
	ginCtx := context.WithValue(ctx, REDIS_SERVICE, redisService)
	RestServe(ginCtx)
	<-ctx.Done()
}

// Make Configuration and Start the RESTFUL API Server
func RestServe(ctx context.Context) {

	if restConfig.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	localIPv4 = utils.LocalIPObtain()
	redisService := ctx.Value("redis-service").(*service.RedisService)

	// Registration of Handler that we will use in the router
	getAllSessionHandler := preparegetAllSessionHandler(redisService)
	getSessionPackets := preparegetSessionPackets(redisService)

	r := gin.Default()
	r.GET("get/session/all", getAllSessionHandler)
	r.GET("get/session/:key", getSessionPackets)
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "no definition of " + c.Request.RequestURI,
		})
	})

	go r.Run(restConfig.Addr)

	fmt.Println("🥳 " + utils.FontSet("Go Gin RESTFUL API Server Start Successfully!"))
	select {
	case <-ctx.Done():
		os.Exit(0)
	default:
	}
}

// preparegetSessionPackets implement the RESTFUL API /get/session/<Key Struct Serd on base64 encoding>
// Its response will be like if everything goes well
// {
// "packets": [
// {
// 		"TTL": 64,
// 		"TcpFlagS": "ACK",
// 		"PayloadExist": false,
// 		"Payload": null,
// 		"PayloadLen": 60,
// 		"SrcIP": "192.168.176.128",
// 		"DstIP": "192.168.176.1",
// 		"SrcPort": 44292,
// 		"DstPort": 1080,
// 		"Timestamp": 1652732483,
// 		"Direction": "Egress"
// 	}]}
func preparegetSessionPackets(redisService *service.RedisService) (fn gin.HandlerFunc) {
	fn = func(c *gin.Context) {
		// Setting Redis query timeout
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

		var direction string // This will indicate the Data Flow direction between client and server
		kk := service.DecodeKey(string(key))
		if kk.DstIP.To4().String() == localIPv4 {
			// Ingress
			direction = "Ingress"
		} else {
			direction = "Egress"
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
					// Not Expected ResultType, inconsistency was detected
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": fmt.Sprintf("inconsitency between expeted Type []redis.Z and received Type %v", notifyMsg.ResultType),
					})
				}

				ZList := notifyMsg.ExecuteResult.([]redis.Z)
				for _, session := range ZList {
					if packet, err := service.DecodeValue(session.Member.(string)); err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					} else { // Value Struct Deserialization is successful
						if !packet.PayloadExist {
							// If payload didn't exists will should keep it nil
							packet.Payload = nil
						}

						value_list = append(value_list, struct {
							*service.Value
							*service.Key
							Timestamp float64
							Direction string
						}{packet, kk, session.Score, direction})
					}
				}
				c.JSON(http.StatusOK, gin.H{
					"msg":  "/get/session/:key response",
					"code": 0,
					"data": value_list,
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

// preparegetALLSessionHandler implement the RESTFUL API /get/all/session
// Its reponse will be like if everything goes well
// {
// "sessions": [
// {
// 	"Key": {
// 		"SrcIP": "192.168.176.128",
// 		"DstIP": "192.168.176.1",
// 		"SrcPort": 44292,
// 		"DstPort": 1080
// 		},
// 	"ID": "Pf-BAwEBA0tleQH_ggABBAEFU3JjSVABCgABBURzdElQAQoAAQdTcmNQb3J0AQYAAQdEc3RQb3J0AQYAAAAX_4IBBMCosIABBMCosAEB_q0EAf4EOAA="
// },
//
// {
// 	"Key": {
// 		"SrcIP": "192.168.176.1",
// 		"DstIP": "192.168.176.128",
// 		"SrcPort": 1080,
// 		"DstPort": 44292
// 		},
// 	"ID": "Pf-BAwEBA0tleQH_ggABBAEFU3JjSVABCgABBURzdElQAQoAAQdTcmNQb3J0AQYAAQdEc3RQb3J0AQYAAAAX_4IBBMCosAEBBMCosIAB_gQ4Af6tBAA="
// 	}
// ]}
func preparegetAllSessionHandler(redisService *service.RedisService) (fn gin.HandlerFunc) {
	fn = func(c *gin.Context) {
		// Setting Redis Query timeout
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
				var key_list []interface{}
				if notifyMsg.ResultType != "[]string" {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": fmt.Sprintf("inconsitency between expeted Type []string and received Type %v", notifyMsg.ResultType),
					})
				}

				sessionList := notifyMsg.ExecuteResult.([]string)
				for _, session := range sessionList {
					key := service.DecodeKey(session)
					key_list = append(key_list, struct {
						Key   *service.Key
						Links interface{}
					}{
						Key: key,
						Links: struct {
							Rel  string
							Href string
						}{
							Rel:  "get specific session info",
							Href: fmt.Sprintf("/get/session/%v", base64.URLEncoding.EncodeToString([]byte(session))),
						},
					})
				}
				c.JSON(http.StatusOK, gin.H{
					"msg":  "/get/session/all response",
					"code": 0,
					"data": key_list,
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
