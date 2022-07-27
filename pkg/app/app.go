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
