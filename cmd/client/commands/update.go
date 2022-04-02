package commands

import (
	"demoserver/api"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update cell <id> <status>",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		///////////////////////////////////////////////////////////////////////////////////////////
		// Parse options
		///////////////////////////////////////////////////////////////////////////////////////////

		// Collect noun (object type) and name
		if args == nil || len(args) < 3 {
			cmd.Help()
			os.Exit(0)
		}

		objectType := args[0]

		if objectType != "cell" {
			config.logger.Fatalf("error: unknown type \"%s\" for \"dmoc %s\"", objectType, cmd.CalledAs())
		}

		objectID := args[1]
		objectStatus := args[2]

		resp, err := config.client.UpdateCell(cmd.Context(), &api.Identifier{Uuid: objectID}, objectStatus)
		if err != nil {
			config.logger.Fatalf("error: %s", err.Error())
		}

		config.logger.Println(resp)
	},
}
