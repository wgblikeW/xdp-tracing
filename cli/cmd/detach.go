/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

const (
	shortDescription_Detach = ""
	longDescription_Detach  = ""
)

// detachCmd represents the detach command
var detachCmd = &cobra.Command{
	Use:   "detach",
	Short: shortDescription_Detach,
	Long:  longDescription_Detach,
	Run:   detachCommandRunFunc,
}

func detachCommandRunFunc(cmd *cobra.Command, args []string) {

}

func init() {
	rootCmd.AddCommand(detachCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// detachCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// detachCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
