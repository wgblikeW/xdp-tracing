package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/p1nant0m/xdp-tracing/handler"
)

type Service interface {
	Conn()
	Serve(taskCh <-chan *AssignTask, notifyCh chan<- *NotifyMsg)
}

type NotifyMsg struct {
	ErrorMsg      error
	ExecuteResult interface{}
	ResultType    string
	Duration      time.Duration
}

type AssignTask struct {
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

func NewRedisService() *RedisService {
	redisService := &RedisService{
		ServiceType: REDIS,
	}
	redisService.MakeNewRedisOptions()
	return redisService
}

func (redisService *RedisService) Serve(taskCh <-chan *AssignTask, notifyCh chan<- *NotifyMsg) {
	for task := range taskCh {
		go func(execTask *AssignTask) {
			start := time.Now()
			result, err := execTask.Task(redisService.Client)
			notifyCh <- &NotifyMsg{
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

type TCP_IPCapturer struct {
	Rules   map[string][]string
	Ctx     context.Context
	Handler func(context.Context, map[string][]string, chan<- *handler.TCP_IP_Handler)
}

func NewTCP_IPCapturer() *TCP_IPCapturer {
	return &TCP_IPCapturer{
		Rules: make(map[string][]string),
		Ctx:   context.Background(),
	}
}

func (capturer *TCP_IPCapturer) Conn() {
	capturer.Handler = handler.StartTCPIPHandler
}

func (capturer *TCP_IPCapturer) Serve(signal chan struct{}, observer chan *TCP_IPCapturer) {

}

func Prototype() {
	ctx := context.Background()
	redisService := NewRedisService()
	redisService.Conn()
	task := func(rdb *redis.Client) (interface{}, error) {
		result, err := rdb.HSet(ctx, "traceid:1234", "SrcIP", "192.168.176.1").Result()
		return result, err
	}

	taskPeriod := &AssignTask{
		Task: func(rdb *redis.Client) (interface{}, error) {
			result, err := rdb.HGetAll(ctx, "traceid:1234").Result()
			return result, err
		},
		ResultType: "map[string]string",
	}

	taskCh := make(chan *AssignTask, 10)
	notifyCh := make(chan *NotifyMsg, 10)
	go redisService.Serve(taskCh, notifyCh)
	var waitGroup sync.WaitGroup

	taskCh <- &AssignTask{
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
