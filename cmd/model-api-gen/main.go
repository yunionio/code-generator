package main

import (
	"os"
	"path/filepath"

	"k8s.io/gengo/args"
	"k8s.io/klog"

	"yunion.io/x/code-generator/pkg/model-api-gen/generators"
)

func main() {
	klog.InitFlags(nil)
	arguments := args.Default()

	// Override defaults.
	arguments.OutputFileBaseName = "zz_generated.model"
	arguments.GoHeaderFilePath = filepath.Join(args.DefaultSourceTree(), "yunion.io/x/code-generator/boilerplate/boilerplate.go.txt")

	if err := arguments.Execute(
		generators.NameSystems(),
		generators.DefaultNameSystem(),
		generators.Packages,
	); err != nil {
		klog.Errorf("Error: %v", err)
		os.Exit(1)
	}
	klog.V(2).Info("Completed successfully.")
}
