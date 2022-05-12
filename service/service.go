package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type Service interface {
	Conn()
	Serve()
}

type notifyMsg struct {
	ErrorMsg      error
	ExecuteResult interface{}
	ResultType    string
	Duration      time.Duration
}

type assignTask struct {
	Task       func(*redis.Client) (interface{}, error)
	ResultType string
}

const (
	REDIS = "redis"
	ETCD  = "etcd"
)

type RedisService struct {
	Options     *redis.Options
	Client      *redis.Client
	ServiceType string
}

func (redisService *RedisService) Serve(taskCh <-chan *assignTask, notifyCh chan<- *notifyMsg) {
	for task := range taskCh {
		go func(execTask *assignTask) {
			start := time.Now()
			result, err := execTask.Task(redisService.Client)
			notifyCh <- &notifyMsg{
				ErrorMsg:      err,
				ExecuteResult: result,
				ResultType:    execTask.ResultType,
				Duration:      time.Since(start),
			}
		}(task)
	}
}

func (redisService *RedisService) Conn() {
	redisService.Client = redis.NewClient(redisService.Options)
}

func NewRedisService() *RedisService {
	redisService := &RedisService{
		ServiceType: REDIS,
	}
	redisService.MakeNewRedisOptions()
	return redisService
}

func Prototype() {
	ctx := context.Background()
	redisService := NewRedisService()
	redisService.Conn()
	task := func(rdb *redis.Client) (interface{}, error) {
		result, err := rdb.HSet(ctx, "traceid:1234", "SrcIP", "192.168.176.1").Result()
		return result, err
	}

	taskPeriod := &assignTask{
		Task: func(rdb *redis.Client) (interface{}, error) {
			result, err := rdb.HGetAll(ctx, "traceid:1234").Result()
			return result, err
		},
		ResultType: "map[string]string",
	}

	taskCh := make(chan *assignTask, 10)
	notifyCh := make(chan *notifyMsg, 10)
	go redisService.Serve(taskCh, notifyCh)
	var waitGroup sync.WaitGroup

	taskCh <- &assignTask{
		Task:       task,
		ResultType: "int64",
	}
	waitGroup.Add(1)
	waitGroup.Add(5) // 5 period task
	go func() {
		for {
			// Task Assigner
			<-time.After(time.Second * 5)
			taskCh <- taskPeriod
		}
	}()

	go func() {
		for feedBack := range notifyCh {
			switch feedBack.ResultType {
			case "int64":
				fmt.Printf("Result:%v ExecuteTime:%v ErrorMsg:%v\n",
					feedBack.ExecuteResult.(int64), feedBack.Duration, feedBack.ErrorMsg)
			case "map[string]string":
				fmt.Printf("Result:%v ExecuteTime:%v ErrorMsg:%v\n",
					feedBack.ExecuteResult.(map[string]string), feedBack.Duration, feedBack.ErrorMsg)
			}
			waitGroup.Done()
		}
	}()
	waitGroup.Wait()
}
