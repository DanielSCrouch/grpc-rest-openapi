package commands

import (
	"demoserver/pkg/client"
	"log"

	"github.com/spf13/cobra"
)

func GetUsage(cmd cobra.Command) {
	cmd.Usage()
}

// Options - Command line option store
type Options struct {
	serverAddr string
	serverPort string
}

// Config - Command line configuration objects
type Config struct {
	logger *log.Logger
	client *client.Client
}

const (
	defaultAddr = "127.0.0.1"
	defaultPort = "8080"
)

var (
	// Config - Command application config
	config Config

	// Options - User command options
	options Options

	// Setup initial command structure
	rootCmd = &cobra.Command{
		Use:   "client",
		Short: "",
		Long:  "",
	}
)

// Execute - Executes the nenadm command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {

	// Setup standard logger to stdout, stderr
	config.logger = log.New(log.Writer(), "", 0)

	// Set global command flags
	rootCmd.PersistentFlags().StringVarP(&options.serverAddr, "address", "a", "", "Server addess")
	rootCmd.PersistentFlags().StringVarP(&options.serverPort, "port", "p", "", "Server port")

	if options.serverAddr == "" {
		options.serverAddr = defaultAddr
	}

	if options.serverPort == "" {
		options.serverPort = defaultPort
	}

	config.client = client.New(options.serverAddr, options.serverPort, config.logger)

}
