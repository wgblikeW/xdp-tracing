// Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/spf13/cobra"
)

type App struct {
	basename    string
	name        string
	description string
	args        cobra.PositionalArgs
	cmd         *cobra.Command
}
