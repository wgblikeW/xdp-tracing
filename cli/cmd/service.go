/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

type serviceFlags struct {
	configPath string
}

var sFlags serviceFlags

const (
	shortDescription_service = ""
	longDescription_service  = ""
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: shortDescription_service,
	Long:  longDescription_service,
	Run:   serviceCommandRunFunc,
}

func serviceCommandRunFunc(cmd *cobra.Command, args []string) {

}

func init() {
	rootCmd.AddCommand(serviceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serviceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serviceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	serviceCmd.Flags().StringVarP(&sFlags.configPath, "conf", "c", "../conf/config.yml", "config file path for service <yml format>")
	serviceCmd.MarkFlagRequired("conf")
}
