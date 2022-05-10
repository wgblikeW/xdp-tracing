/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/p1nant0m/xdp-tracing/bpf"
	"github.com/p1nant0m/xdp-tracing/bpf/errors"
	"github.com/p1nant0m/xdp-tracing/config"
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

type attachMode struct {
	// XDP_FLAGS_MODES [https://github1s.com/libbpf/libbpf/blob/93c570ca4b251415b72ef24e17d7a93ca61a9d42/include/uapi/linux/if_link.h#L1185]
	SKB_MODE    bool
	NATIVE_MODE bool
	HW_MODE     bool
}

type attachFlags struct {
	FileName string
	*attachMode
}

var aFlags = &attachFlags{
	attachMode: &attachMode{},
}

func attachCommandRunFunc(cmd *cobra.Command, args []string) {
	gFlags, _ := getGlobalFlags(cmd)

	cfg := config.NewXDPConfig()
	// reconfigure the XDP Config
	cfg.Ifname = gFlags.DevName
	cfg.Filename = aFlags.FileName
	cfg.Xdp_flags = parseAttachMode(aFlags.attachMode)

	// Convert Go type to C type and call the function
	errCode := bpf.Attach_bpf_prog_to_if(cfg)

	fmt.Println(errors.GetErrorString(errCode))
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
	attachCmd.PersistentFlags().BoolVarP(&aFlags.SKB_MODE, "sk-mode", "S", false, "Install XDP program in SKB (AKA generic) mode")
	attachCmd.PersistentFlags().BoolVarP(&aFlags.NATIVE_MODE, "native-mode", "N", false, "Install XDP program in native mode")
	attachCmd.PersistentFlags().BoolVarP(&aFlags.HW_MODE, "hw-mode", "H", false, "Install XDP program in hw mode")
	attachCmd.PersistentFlags().StringVarP(&aFlags.FileName, "file-name", "f", "../../bpf/xdp_proxy_kern.o", "binary eBPF program path")
	attachCmd.MarkPersistentFlagRequired("file-name")
}

// parseAttachMode return the XDP Attaching Mode Flags based on CLI input
func parseAttachMode(attachModeSel *attachMode) (flags uint32) {
	flags = config.XDP_FLAGS_UPDATE_IF_NOEXIST
	if attachModeSel.SKB_MODE {
		flags |= config.XDP_FLAGS_SKB_MODE
		goto out
	}
	if attachModeSel.NATIVE_MODE {
		flags |= config.XDP_FLAGS_DRV_MODE
		goto out
	}
	if attachModeSel.HW_MODE {
		flags |= config.XDP_FLAGS_HW_MODE
		goto out
	}
	flags |= config.XDP_FLAGS_SKB_MODE // default setting of Attach Mode
out:
	return flags
}
