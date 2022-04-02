package main

import (
	"context"
	"demoserver/pkg/server"
	"log"
	"os"
)

func main() {

	serverClosedCh := make(chan struct{})
	logger := log.Default()

	//////////////////////////////////////////////////////////////////
	// Get ENV variables
	//////////////////////////////////////////////////////////////////

	// Set server address
	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		serverAddr = "0.0.0.0"
	}
	logger.Printf("SERVER_ADDR: %s", serverAddr)

	// Set server port
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}
	logger.Printf("SERVER_PORT: %s", serverPort)

	//////////////////////////////////////////////////////////////////
	// Start gRPC server
	//////////////////////////////////////////////////////////////////

	server := server.New(serverAddr, serverPort, logger, &serverClosedCh)
	go server.Serve(context.Background())

	go server.HTTPReverseProxy()

	//////////////////////////////////////////////////////////////////
	// Check for fatal errors
	//////////////////////////////////////////////////////////////////

	<-serverClosedCh
	os.Exit(1)
}
