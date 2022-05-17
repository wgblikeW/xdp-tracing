/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/p1nant0m/xdp-tracing/handler"
	"github.com/p1nant0m/xdp-tracing/handler/utils"
	"github.com/p1nant0m/xdp-tracing/service"
	"github.com/p1nant0m/xdp-tracing/service/rest"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type serviceFlags struct {
	configPath string
}

var sFlags serviceFlags

const (
	shortDescription_service = ""
	longDescription_service  = ""
	DEBUG_ENABLE             = false
)

func init() {
	if DEBUG_ENABLE {
		logrus.SetLevel(logrus.DebugLevel)
	}
	rootCmd.AddCommand(serviceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serviceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serviceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	serviceCmd.PersistentFlags().StringVarP(&sFlags.configPath, "conf", "c", "../conf/config.yml", "config file path for service <yml format>")
	serviceCmd.MarkFlagRequired("conf")
}

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: shortDescription_service,
	Long:  longDescription_service,
	Run:   serviceCommandRunFunc,
}

// serviceCommandRunFunc Runs the main logic of CLI Command "service"
func serviceCommandRunFunc(cmd *cobra.Command, args []string) {
	logrus.Debug("In serviceCommandRunFunc:66")

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

	// Reading Configuaration from Config file
	filePath, _ := cmd.PersistentFlags().GetString("conf")
	err := service.ReadAndParseConfig(filePath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Startup Redis Service
	redisService := startRedisComponet(ctx)

	// StartUp Packets Capture
	observeCh := startPacketsCap(ctx)

	// Making Data Flow From local Capturer to remote RedisDB
	redisService.Register("capturer") // capturer need to use Redis Service, so it need to regist first
	streamFlow_Cap2Rdb(ctx, redisService, observeCh)

	// Start Rest Server
	ginCtx := context.WithValue(ctx, "redis-service", redisService)
	rest.RestServe(ginCtx)

	// Make Registration in ETCD
	startEtcdComponet(ctx)
	fmt.Println("🥳 " + utils.FontSet("All Services Start successfully! Enjoy your Days!"))
	<-ctx.Done()
}

func startEtcdComponet(ctx context.Context) *service.EtcdService {
	var etcdService service.Service = service.NewEtcdService(ctx)
	if err := etcdService.Conn(); err != nil {
		log.Fatal(err.Error())
	}

	etcdService.Serve()
	return etcdService.(*service.EtcdService)
}

// streamFlow_Cap2Rdb make data flow from local capturer to Redis
func streamFlow_Cap2Rdb(ctx context.Context,
	redisService *service.RedisService, packetCh <-chan *handler.TCP_IP_Handler) {
	redisNotifyCh, err := redisService.RetrieveChannel("capturer")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// this Goroutine handles the response from Redis via redisNotifyCh
	go func() {
		for notifyMsg := range redisNotifyCh {
			select {
			case <-ctx.Done():
				return
			default:
				// Previous Task Finished handle the remainning process
				// TODO: Retry if something errors happen
				handleRespFromRdb(notifyMsg)
			}
		}
	}()

	// this Goroutine Records the new filtered Packets to Redis
	go func() {
		for packet := range packetCh {
			select {
			case <-ctx.Done():
				return
			default:
				logrus.Debug("new packet arrives Packets:%v", packet)
				// packet that satisfied the rules arrive,
				// new task should be assgined to Redis Client
				taskFunc, resultType := newRecordTask(ctx, packet)
				redisService.TaskAssign(taskFunc, resultType, "capturer")
			}
		}
	}()

}

// newRecordTask construct the Redis Task to make record of arriving packet
func newRecordTask(ctx context.Context, packet *handler.TCP_IP_Handler) (func(rdb *redis.Client) (interface{}, error), string) {
	key := &service.Key{
		SrcIP:   packet.SrcIP,
		DstIP:   packet.DstIP,
		SrcPort: packet.SrcPort,
		DstPort: packet.DstPort,
	}

	value := &service.Value{
		TTL:          packet.TTL,
		TcpFlagS:     packet.TcpFlagsS,
		PayloadExist: packet.PayloadExist,
		PayloadMeta:  packet.PayloadMeta,
	}

	// serialize the Key struct and Value struct
	keyS, valueS := service.EncodeSession(key, value)

	// using for sorted list score
	timeT, _ := time.Parse("2006-01-02 15:04:05.999999999", packet.Timestamp)
	timeF := float64(timeT.Unix())

	taskFunc := func(rdb *redis.Client) (interface{}, error) {
		cmds, err := rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.ZAdd(ctx, keyS, &redis.Z{Score: timeF, Member: valueS})
			pipe.SAdd(ctx, "sessions", keyS)
			return nil
		})
		return cmds, err
	}

	return taskFunc, "[]redis.Cmder"

}

// handleRespFromRdb process the response from Redis Server after we submit the Task to the
// server
func handleRespFromRdb(resp *service.NotifyMsg) {
	// fmt.Printf("Client:%v Duration:%v ErrorMsg:%v ExecuteReuslt:%v ResultType:%v", resp.Client, resp.Duration, resp.ErrorMsg,
	// 	resp.ExecuteResult, resp.ResultType)
}

func startPacketsCap(ctx context.Context) <-chan *handler.TCP_IP_Handler {
	// Create New Instance of TCP_IPCapturer
	capturer := service.NewTCP_IPCapturer(ctx)

	capturer.MakeNewRules()
	capturer.Conn()

	observeCh := make(chan *handler.TCP_IP_Handler)
	capturer.Serve(observeCh)
	return observeCh
}

func startRedisComponet(ctx context.Context) *service.RedisService {
	// Setup Redis Service
	var redisServe service.Service = service.NewRedisService(ctx)

	redisServe.Conn()
	go redisServe.Serve()

	return redisServe.(*service.RedisService)
}
