/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

const (
	longDescription_Attach  = ""
	shortDescription_Attach = ""
)

// attachCmd represents the attach command
var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: longDescription_Attach,
	Long:  shortDescription_Attach,
	Run:   attachCommandRunFunc,
}

func attachCommandRunFunc(cmd *cobra.Command, args []string) {
	watcher := make(chan os.Signal, 1)
	signal.Notify(watcher, os.Interrupt, syscall.SIGTERM)
	ctx, cancelFunc := context.WithCancel(context.Background())

}

func init() {
	rootCmd.AddCommand(attachCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// attachCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// attachCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
