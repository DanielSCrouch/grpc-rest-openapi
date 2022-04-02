package commands

import (
	"demoserver/api"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get cell <id>",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		///////////////////////////////////////////////////////////////////////////////////////////
		// Parse options
		///////////////////////////////////////////////////////////////////////////////////////////

		// Collect noun (object type) and name
		if args == nil || len(args) < 2 {
			cmd.Help()
			os.Exit(0)
		}

		objectType := args[0]
		objectID := args[1]

		if objectType != "cell" {
			config.logger.Fatalf("error: unknown type \"%s\" for \"dmoc %s\"", objectType, cmd.CalledAs())
		}

		cell, err := config.client.GetCell(cmd.Context(), &api.IdMessage{Uuid: objectID})
		if err != nil {
			config.logger.Fatalf("error: %s", err.Error())
		}

		config.logger.Println(cell)
	},
}
