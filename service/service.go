package service

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/p1nant0m/xdp-tracing/handler"
	"github.com/sirupsen/logrus"
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

// ---------------------------------------------------- Redis Service ------------------------------

type RedisService struct {
	Options     *redis.Options
	Client      *redis.Client
	Ctx         context.Context
	ServiceType string
}

func NewRedisService(ctx context.Context) *RedisService {
	redisService := &RedisService{
		ServiceType: REDIS,
		Ctx:         ctx,
	}
	redisService.MakeNewRedisOptions()
	return redisService
}

func (redisService *RedisService) Conn() {
	redisService.Client = redis.NewClient(redisService.Options)
}

// Serve starts a goroutine to receive "Redis Command Task" from task channel
// and submit the task to the redis Client. The reply from the redis server will
// be warpped and the Task producer will be informed via notify Channel
func (redisService *RedisService) Serve(taskCh <-chan *AssignTask, notifyCh chan<- *NotifyMsg) {
	logrus.Debug("In redisService.Serve:51")
	for task := range taskCh {
		select {
		case <-redisService.Ctx.Done():
			//TODO: waiting for all requesets are properly process
			return
		default:
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
}

// ---------------------------------------------------- Redis Service ------------------------------

//---------------------------------------------------- TCP_IPCapturer ------------------------------

type TCP_IPCapturer struct {
	Rules   map[string][]string
	Ctx     context.Context
	Handler func(context.Context, map[string][]string, chan<- *handler.TCP_IP_Handler)
}

func NewTCP_IPCapturer(ctx context.Context) *TCP_IPCapturer {
	return &TCP_IPCapturer{
		Rules: make(map[string][]string),
		Ctx:   ctx,
	}
}

func (capturer *TCP_IPCapturer) Conn() {
	capturer.Handler = handler.StartTCPIPHandler
}

func (capturer *TCP_IPCapturer) Serve(observer chan<- *handler.TCP_IP_Handler) {
	go func() {
		capturer.Handler(capturer.Ctx, capturer.Rules, observer)
		// if everything goes well, it will not reach the block below
	}()
}

//---------------------------------------------------- TCP_IPCapturer ------------------------------
