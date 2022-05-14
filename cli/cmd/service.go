/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/gopacket/layers"
	"github.com/p1nant0m/xdp-tracing/handler"
	"github.com/p1nant0m/xdp-tracing/service"
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
	DEBUG_ENABLE             = true
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

func serviceCommandRunFunc(cmd *cobra.Command, args []string) {
	logrus.Debug("In serviceCommandRunFunc:66")

	// Setup notifier and Make Configuration of All Services
	ctx, cancel := context.WithCancel(context.Background())
	watcher := make(chan os.Signal, 1)
	signal.Notify(watcher, os.Interrupt, syscall.SIGTERM)
	defer cancel()

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
	redisTaskCh, redisNotifyCh := startRedisComponet(ctx)

	// StartUp Packets Capture
	observeCh := startPacketsCap(ctx)

	streamFlow_Cap2Rdb(ctx, redisTaskCh, redisNotifyCh, observeCh)
	<-ctx.Done()
}

func streamFlow_Cap2Rdb(ctx context.Context, redisTaskCh chan<- *service.AssignTask,
	redisNotifyCh <-chan *service.NotifyMsg, packetCh <-chan *handler.TCP_IP_Handler) {

	// this Goroutine handles the response from Redis via redisNotifyCh
	go func() {
		logrus.Debug("Goroutine handles the response from Redis via redisNotifyCh")
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
		logrus.Debug("this Goroutine Records the new filtered Packets to Redis")
		for packet := range packetCh {
			select {
			case <-ctx.Done():
				return
			default:
				logrus.Debug("new packet arrives Packets:%v", packet)
				// packet that satisfied the rules arrive,
				// new task should be assgined to Redis Client
				redisTaskCh <- newRecordTask(ctx, packet)
			}
		}
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

func newRecordTask(ctx context.Context, packet *handler.TCP_IP_Handler) *service.AssignTask {
	key := &Key{
		SrcIP:   packet.SrcIP,
		DstIP:   packet.DstIP,
		SrcPort: packet.SrcPort,
		DstPort: packet.DstPort,
	}

	value := &Value{
		TTL:          packet.TTL,
		TcpFlagS:     packet.TcpFlagsS,
		PayloadExist: packet.PayloadExist,
		PayloadMeta:  packet.PayloadMeta,
	}

	// serialize the Key struct as Sorted List Key
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(*key)
	keyS := buf.String()

	// serialize the Value strcut as the store element
	enc.Encode(*value)
	valueS := buf.String()

	// using for sorted list score
	timeT, _ := time.Parse("2006-01-02 15:04:05.999999999", packet.Timestamp)
	timeF := float64(timeT.Unix())

	taskFunc := func(rdb *redis.Client) (interface{}, error) {
		cmds, err := rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.ZAdd(ctx, keyS, &redis.Z{Score: timeF, Member: valueS})
			pipe.LPush(ctx, "packet:tcpip", keyS)
			return nil
		})
		return cmds, err
	}

	return &service.AssignTask{
		Task:       taskFunc,
		ResultType: "[]redis.Cmder",
	}

}

func handleRespFromRdb(resp *service.NotifyMsg) {
	fmt.Printf("Duration:%v ErrorMsg:%v ExecuteResult:%v ResultType:%v",
		resp.Duration, resp.ErrorMsg,
		resp.ExecuteResult, resp.ResultType)
}

func startPacketsCap(ctx context.Context) <-chan *handler.TCP_IP_Handler {
	logrus.Debug("In startPacketsCap:210")
	// Create New Instance of TCP_IPCapturer
	capturer := service.NewTCP_IPCapturer(ctx)

	capturer.MakeNewRules()
	capturer.Conn()

	observeCh := make(chan *handler.TCP_IP_Handler)
	capturer.Serve(observeCh)
	return observeCh
}

func startRedisComponet(ctx context.Context) (chan<- *service.AssignTask, <-chan *service.NotifyMsg) {
	logrus.Debug("In startRedisComponet:222")
	// Setup Redis Service
	var redisServe service.Service = service.NewRedisService(ctx)
	redisServe.Conn()

	taskCh := make(chan *service.AssignTask, 10)
	notifyCh := make(chan *service.NotifyMsg, 10)
	go redisServe.Serve(taskCh, notifyCh)
	return taskCh, notifyCh
}
