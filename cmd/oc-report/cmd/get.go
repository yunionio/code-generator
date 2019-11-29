package cmd

import (
	"github.com/spf13/cobra"
)

type getOptions struct {
}

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "display one or many types",
	}
	cmd.AddCommand(newGetCmd())
	return cmd
}

func newGetKind() *cobra.Command {
	return nil
}
