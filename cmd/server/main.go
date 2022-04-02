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
	ctx := context.Background()

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
		serverPort = "8081"
	}
	logger.Printf("SERVER_PORT: %s", serverPort)

	// Set proxy server port
	proxyPort := os.Getenv("PROXY_SERVER_PORT")
	if proxyPort == "" {
		proxyPort = "8080"
	}
	logger.Printf("PROXY_SERVER_PORT: %s", proxyPort)

	//////////////////////////////////////////////////////////////////
	// Start gRPC server
	//////////////////////////////////////////////////////////////////

	server := server.New(serverAddr, serverPort, proxyPort, logger, &serverClosedCh)
	go server.Serve(ctx)
	go server.ServeReverseProxy(ctx)

	//////////////////////////////////////////////////////////////////
	// Check for fatal errors
	//////////////////////////////////////////////////////////////////

	<-serverClosedCh
	os.Exit(1)
}
