/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/p1nant0m/xdp-tracing/service"
	"github.com/p1nant0m/xdp-tracing/service/rest"
	"github.com/spf13/cobra"
)

type restserviceFlags struct {
	configPath string
}

var rFlags restserviceFlags
var restConfig *service.RestConfig

const (
	shortDescription_RestServe = ""
	longDescription_RestServe  = ""
)

// restserveCmd represents the restserve command
var restserveCmd = &cobra.Command{
	Use:   "restserve",
	Short: shortDescription_RestServe,
	Long:  longDescription_RestServe,
	Run:   restServeCommandRunFunc,
}

func restServeCommandRunFunc(cmd *cobra.Command, args []string) {
	rest.RunRestServer(rFlags.configPath)
}

func init() {
	rootCmd.AddCommand(restserveCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// restserveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// restserveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	restserveCmd.PersistentFlags().StringVarP(&rFlags.configPath, "conf", "c", "../conf/config.yml", "config file path for service <yml format>")
	restserveCmd.MarkPersistentFlagRequired("conf")
}
