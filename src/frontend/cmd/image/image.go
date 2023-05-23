package image

import (
	"podkit/frontend/cmd/image/ls"
	"podkit/frontend/cmd/image/start"

	"github.com/spf13/cobra"
)

func init() {
	ImageCmd.AddCommand(ls.LSCmd)
	ImageCmd.AddCommand(start.StartCmd)
}

var ImageCmd = &cobra.Command{
	Use:   "image",
	Short: "ls command is used to manage images",
}
