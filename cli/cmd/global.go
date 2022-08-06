// Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cmd

import "github.com/spf13/cobra"

var globalFlags = GlobalFlags{}

// GlobalFlags are flags that defined globally and are inherited to all sub-commands.
type GlobalFlags struct {
	DevName string
}

func getGlobalFlags(command *cobra.Command) (conf GlobalFlags, err error) {
	conf.DevName, err = command.Flags().GetString("dev")
	if err != nil {
		return
	}
	return
}
