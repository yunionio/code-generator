package main

import (
	goflag "flag"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"yunion.io/x/code-generator/cmd/swagger-serve/cmd"
)

func main() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
