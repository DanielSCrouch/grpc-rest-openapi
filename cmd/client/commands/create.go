package commands

import (
	"demoserver/api"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create cell <id> <status>",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		///////////////////////////////////////////////////////////////////////////////////////////
		// Parse options
		///////////////////////////////////////////////////////////////////////////////////////////

		// Collect noun (object type) and name
		if args == nil || len(args) < 1 {
			cmd.Help()
			os.Exit(0)
		}

		objectType := args[0]

		if objectType != "cell" {
			config.logger.Fatalf("error: unknown type \"%s\" for \"dmoc %s\"", objectType, cmd.CalledAs())
		}

		var objectID string
		var objectStatus string
		if len(args) > 1 {
			objectID = args[1]
		}

		if len(args) > 2 {
			objectStatus = args[2]
		} else {
			objectStatus = "offline"
		}

		cell := &api.Cell{Identity: &api.Identifier{Uuid: objectID}, Status: objectStatus}
		resp, err := config.client.CreateCell(cmd.Context(), cell)
		if err != nil {
			config.logger.Fatalf("error: %s", err.Error())
		}

		config.logger.Println(resp)
	},
}
