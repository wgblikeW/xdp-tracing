/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/p1nant0m/xdp-tracing/bpf"
	"github.com/p1nant0m/xdp-tracing/bpf/errors"
	"github.com/spf13/cobra"
)

const (
	shortDescription_Detach = ""
	longDescription_Detach  = ""
)

type detachFlags struct {
	prog_id int
}

var dFlags = &detachFlags{
	prog_id: 0,
}

// detachCmd represents the detach command
var detachCmd = &cobra.Command{
	Use:   "detach",
	Short: shortDescription_Detach,
	Long:  longDescription_Detach,
	Run:   detachCommandRunFunc,
}

func detachCommandRunFunc(cmd *cobra.Command, args []string) {
	gFlags, _ := getGlobalFlags(cmd)
	errCode := bpf.Warp_do_detach(gFlags.DevName, dFlags.prog_id)

	fmt.Println(errors.GetErrorString(errCode))
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
	detachCmd.PersistentFlags().IntVarP(&dFlags.prog_id, "prog-id", "p", 0, "BPF program Index that need to detach <using bpftool prog to check>")
	detachCmd.PersistentFlags().StringVarP(&globalFlags.DevName, "dev", "d", "eth0", "Operate on device <ifname>")
	detachCmd.MarkFlagRequired("dev")
	detachCmd.MarkFlagRequired("prog-id")
}
