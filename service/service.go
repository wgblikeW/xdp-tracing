package service

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/gopacket/layers"
	"github.com/p1nant0m/xdp-tracing/handler"
)

type Service interface {
	Conn()
	Serve()
}

type NotifyMsg struct {
	ErrorMsg      error
	ExecuteResult interface{}
	ResultType    string
	Duration      time.Duration
	Client        string
}

type AssignTask struct {
	Task       func(*redis.Client) (interface{}, error)
	ResultType string
	Client     string
}

const (
	REDIS = "redis"
	ETCD  = "etcd"
)

// ---------------------------------------------------- Redis Service ------------------------------

type RedisService struct {
	Options     *redis.Options
	RDClient    *redis.Client
	Ctx         context.Context
	TaskCh      chan *AssignTask
	NotifyCh    chan *NotifyMsg
	sClients    map[string]chan *NotifyMsg
	ServiceType string
}

func NewRedisService(ctx context.Context) *RedisService {
	redisService := &RedisService{
		ServiceType: REDIS,
		Ctx:         ctx,
		TaskCh:      make(chan *AssignTask),
		NotifyCh:    make(chan *NotifyMsg),
		sClients:    make(map[string]chan *NotifyMsg),
	}
	redisService.MakeNewRedisOptions()
	return redisService
}

func (redisService *RedisService) Conn() {
	redisService.RDClient = redis.NewClient(redisService.Options)
}

func (redisService *RedisService) Register(client string) {
	ticket := make(chan *NotifyMsg)
	redisService.sClients[client] = ticket
}

func (redisService *RedisService) Destory(client string) {
	close(redisService.sClients[client])
	delete(redisService.sClients, client)
}

func (redisService *RedisService) RetrieveChannel(client string) (<-chan *NotifyMsg, error) {
	ch, ok := redisService.sClients[client]
	if !ok {
		return nil, errors.New("regist for the service before using")
	}
	return ch, nil
}

// Serve starts a goroutine to receive "Redis Command Task" from task channel
// and submit the task to the redis Client. The reply from the redis server will
// be warpped and the Task producer will be informed via notify Channel
func (redisService *RedisService) Serve() {
	go redisService.responseHandler()
	for task := range redisService.TaskCh {
		select {
		case <-redisService.Ctx.Done():
			//TODO: waiting for all requesets are properly process
			return
		default:
			go func(execTask *AssignTask) {
				start := time.Now()
				result, err := execTask.Task(redisService.RDClient)
				redisService.NotifyCh <- &NotifyMsg{
					ErrorMsg:      err,
					ExecuteResult: result,
					ResultType:    execTask.ResultType,
					Duration:      time.Since(start),
					Client:        execTask.Client,
				}
			}(task)
		}
	}
}

func (redisService *RedisService) TaskAssign(taskFunc func(*redis.Client) (interface{}, error),
	resultType string, client string) {

	redisService.TaskCh <- &AssignTask{
		Task:       taskFunc,
		ResultType: resultType,
		Client:     client,
	}
}

func (redisService *RedisService) responseHandler() {
	for notifyMsg := range redisService.NotifyCh {
		select {
		case <-redisService.Ctx.Done():
			return
		default:
			redisService.sClients[notifyMsg.Client] <- notifyMsg
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

type Key struct {
	SrcIP   net.IP
	DstIP   net.IP
	SrcPort layers.TCPPort
	DstPort layers.TCPPort
}

type Value struct {
	TTL          uint8
	TcpFlagS     string
	PayloadExist bool
	*handler.PayloadMeta
}

func DecodeKey(keySerdString string) *Key {
	var buf bytes.Buffer
	key := &Key{}
	dec := gob.NewDecoder(&buf)
	buf.WriteString(keySerdString)
	dec.Decode(key)
	return key
}

func DecodeValue(valueSerdString string) (*Value, error) {
	var buf bytes.Buffer
	var value Value
	dec := gob.NewDecoder(&buf)
	buf.WriteString(valueSerdString)
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	return &value, nil
}

func DecodeSession(keySerdString string, valueSerdString string) (*Key, *Value) {
	var buf bytes.Buffer
	key, value := &Key{}, &Value{}
	dec := gob.NewDecoder(&buf)
	buf.WriteString(keySerdString)
	dec.Decode(key)

	buf.WriteString(valueSerdString)
	dec.Decode(value)
	return key, value
}

func EncodeSession(key *Key, value *Value) (string, string) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		panic(err.Error())
	}
	keyString := buf.String()

	buf.Reset()

	err = enc.Encode(value)
	if err != nil {
		panic(err.Error())
	}
	valueString := buf.String()

	return keyString, valueString
}

//---------------------------------------------------- TCP_IPCapturer ------------------------------
