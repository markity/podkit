package main

// podkit

import (
	"fmt"
	"os"
	"podkit/frontend/cmd/container"
	"podkit/frontend/cmd/image"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:               "podkit",
	Short:             "Podkit is an easy linux container kit.",
	Long:              "Podkit helps you better understand the mechanism of docker. It provides main functions to get you understand how docker works. ",
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("Root permission is required")
		return
	}
	// podkit image ls
	// podkit image start ubuntu22.04
	RootCmd.AddCommand(image.ImageCmd)
	RootCmd.AddCommand(container.ContainerCmd)
	RootCmd.Execute()
}
