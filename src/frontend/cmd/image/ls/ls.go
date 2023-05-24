package ls

import (
	"fmt"
	"podkit/frontend/json_struct"
	"podkit/frontend/tools"

	"github.com/spf13/cobra"
)

var LSCmd = &cobra.Command{
	Use:   "ls",
	Short: "list all availabel images",
	Args:  cobra.MatchAll(cobra.ExactArgs(0)),
	Run: func(cmd *cobra.Command, args []string) {
		flock := tools.FlockManager{}
		err := flock.Init("/var/lib/podkit/lock")
		if err != nil {
			panic(err)
		}
		err = flock.Lock()
		if err != nil {
			panic(err)
		}
		defer flock.Release()

		s := json_struct.ImageInfoStruct{}
		s.ParseFromFile("/var/lib/podkit/images/image_info.json")
		if err != nil {
			panic(err)
		}

		for k := range s.ImageTarFilename {
			fmt.Println(k)
		}
	},
}
