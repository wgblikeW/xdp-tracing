/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const (
	cliName             = "xdp-tracing"
	cliShortDescription = "./xdp-tracing [options]"
	cliLongDescription  = ""
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   cliName,
	Short: cliShortDescription,
	Long:  cliLongDescription,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	var globalFlags = GlobalFlags{}
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.xdp-tracing.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVarP(&globalFlags.DevName, "dev", "d", "eth0", "Operate on device <ifname>")
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.SKB_MODE, "sk-mode", "S", false, "Install XDP program in SKB (AKA generic) mode")
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.NATIVE_MODE, "native-mode", "N", false, "Install XDP program in native mode")
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.HW_MODE, "hw-mode", "H", false, "Install XDP program in hw mode")
}
