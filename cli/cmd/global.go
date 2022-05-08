package cmd

import "github.com/spf13/cobra"

// GlobalFlags are flags that defined globally and are inherited to all sub-commands.
type GlobalFlags struct {
	DevName string

	// XDP_FLAGS_MODES [https://github1s.com/libbpf/libbpf/blob/93c570ca4b251415b72ef24e17d7a93ca61a9d42/include/uapi/linux/if_link.h#L1185]
	SKB_MODE    bool
	NATIVE_MODE bool
	HW_MODE     bool
}

func getGlobalConf(command *cobra.Command) (conf GlobalFlags, err error) {
	conf.DevName, err = command.Flags().GetString("dev")
	if err != nil {
		return
	}

	conf.SKB_MODE, err = command.Flags().GetBool("sk-mode")
	if err != nil {
		return
	}

	conf.NATIVE_MODE, err = command.Flags().GetBool("native-mode")
	if err != nil {
		return
	}

	conf.HW_MODE, err = command.Flags().GetBool("hw-mode")
	if err != nil {
		return
	}

	return
}
