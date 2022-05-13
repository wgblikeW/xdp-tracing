/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/p1nant0m/xdp-tracing/handler"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rules = make(map[string][]string)

const (
	longDescription_capture  = ""
	shortDescription_capture = ""
)

type captureFlags struct {
	SrcIP   []string
	DstIP   []string
	SrcPort []string
	DstPort []string
}

var cFlags = &captureFlags{}

// captureCmd represents the capture command
var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: shortDescription_capture,
	Long:  longDescription_capture,
	Run:   captureCommandRunFunc,
}

func makeRulesWithFlags(flags *pflag.FlagSet) {
	if value, err := flags.GetStringArray("src-ip"); err == nil && flags.Changed("src-ip") {
		rules["SrcIP"] = value
	}
	if value, err := flags.GetStringArray("dst-ip"); err == nil && flags.Changed("dst-ip") {
		rules["DstIP"] = value
	}
	if value, err := flags.GetStringArray("src-port"); err == nil && flags.Changed("src-port") {
		rules["SrcPort"] = value
	}
	if value, err := flags.GetStringArray("dst-port"); err == nil && flags.Changed("dst-port") {
		rules["dst-port"] = value
	}
}

func captureCommandRunFunc(cmd *cobra.Command, args []string) {
	watcher := make(chan os.Signal, 1)
	// stopCh := make(chan struct{})
	observerCh := make(chan *handler.TCP_IP_Handler, 100)
	signal.Notify(watcher, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-watcher
		// OS Signal Catched, exit the program gracefully
		cancel()
	}()

	makeRulesWithFlags(cmd.PersistentFlags())
	go func() {
		handler.StartTCPIPHandler(ctx, rules, observerCh)
		// if everything goes well, it will not reach the block below
		cancel()
	}()

	// display Captured Packets
	go func(ctx context.Context) {
		for packet := range observerCh {

			fmt.Printf("[%s] %s:%d -> %s:%d [%s] TTL:%d\n", packet.Timestamp, packet.SrcIP, packet.SrcPort, packet.DstIP, packet.DstPort, packet.TcpFlagsS, packet.TTL)
			if packet.PayloadExist {
				fmt.Println(hex.Dump(*packet.Payload))
			}

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}(ctx)

	// select {
	// case <-watcher:
	// 	fmt.Println("\nUser Terminates the World")
	// case <-stopCh:
	// }
	<-ctx.Done()
}

func init() {
	rootCmd.AddCommand(captureCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// captureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// captureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	captureCmd.PersistentFlags().StringArrayVarP(&cFlags.SrcIP, "src-ip", "s", []string{}, "filter Source IPv4 Address (format xxx.xxx.xxx.xxx)")
	captureCmd.PersistentFlags().StringArrayVarP(&cFlags.DstIP, "dst-ip", "t", []string{}, "filter Destination IPv4 Address (format xxx.xxx.xxx.xxx)")
	captureCmd.PersistentFlags().StringArrayVarP(&cFlags.SrcPort, "src-port", "p", []string{}, "filter Source Port")
	captureCmd.PersistentFlags().StringArrayVarP(&cFlags.DstPort, "dst-port", "o", []string{}, "filter Destination Port")
}
