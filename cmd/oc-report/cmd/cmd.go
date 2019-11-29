package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "oc-report",
		Short: "reporter for onecloud project",
	}
	cmds.AddCommand(newGetCmd())
	return cmds
}
