package client

import (
	"context"
	"demoserver/api"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	ctx        context.Context
	serverAddr string
	serverPort string
	logger     *log.Logger
}

func New(serverAddr, serverPort string, logger *log.Logger) (client *Client) {
	return &Client{
		ctx:        context.TODO(),
		serverAddr: serverAddr,
		serverPort: serverPort,
		logger:     logger,
	}
}

// connect - Returns gRPC client connection between Client and Server
func (c *Client) connect(ctx context.Context) (connClient *api.OrchestratorServiceClient, close context.CancelFunc, err error) {

	c.logger.Printf("connecting to server on %s:%s", c.serverAddr, c.serverPort)
	// Create gRPC client
	var opts []grpc.DialOption
	credentials := insecure.NewCredentials()
	opts = append(opts, grpc.WithTransportCredentials(credentials))
	opts = append(opts, grpc.WithBlock())
	ctx, close = context.WithTimeout(ctx, 5*time.Second)
	grpcConn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%s", c.serverAddr, c.serverPort), opts...)
	if err != nil {
		return nil, nil, err
	}

	grpcClient := api.NewOrchestratorServiceClient(grpcConn)

	return &grpcClient, close, nil
}

// GetCell - Returns Cell matched by ID
func (c *Client) GetCell(ctx context.Context, uuid *api.IdMessage) (cell *api.Cell, err error) {

	c.logger.Println("connecting...")
	connClient, close, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer close()

	c.logger.Println("getting cells...")

	cell, err = (*connClient).GetCell(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return cell, nil
}
