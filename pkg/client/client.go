package client

import (
	"context"
	"demoserver/api"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var grpcHeaders = metadata.New(map[string]string{
	"Content-Type":         "application/grpc",
	"grpc-accept-encoding": "identity,deflate,gzip",
	"scheme":               "http",
})

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
func (c *Client) connect(ctx context.Context) (connClient *api.CellServiceClient, close context.CancelFunc, err error) {

	c.logger.Printf("connecting to server on %s:%s", c.serverAddr, c.serverPort)
	// Create gRPC client
	var opts []grpc.DialOption
	credentials := insecure.NewCredentials()
	opts = append(opts, grpc.WithTransportCredentials(credentials))
	ctx = metadata.NewOutgoingContext(ctx, grpcHeaders)
	opts = append(opts, grpc.WithBlock())
	ctx, close = context.WithTimeout(ctx, 5*time.Second)
	grpcConn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%s", c.serverAddr, c.serverPort), opts...)
	if err != nil {
		return nil, nil, err
	}

	grpcClient := api.NewCellServiceClient(grpcConn)

	return &grpcClient, close, nil
}

// CreateCell - Calls the Server to create a Cell
func (c *Client) CreateCell(ctx context.Context, cell *api.Cell) (resp *api.Cell, err error) {

	connClient, close, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer close()

	c.logger.Println("creating cell...")
	var headerResp metadata.MD
	req := api.CreateCellRequest{Cell: cell}
	resp, err = (*connClient).CreateCell(ctx, &req, grpc.Header(&headerResp))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetCell - Returns a Cell matched by ID
func (c *Client) GetCell(ctx context.Context, id *api.Identifier) (cell *api.Cell, err error) {

	connClient, close, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer close()

	c.logger.Println("getting cell...")
	resp, err := (*connClient).GetCell(ctx, &api.GetCellRequest{Identity: id})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// UpdateCellStatus - Updates a Cell's statuus matched by ID
func (c *Client) UpdateCell(ctx context.Context, id *api.Identifier, status string) (cell *api.Cell, err error) {

	connClient, close, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer close()

	cell, err = c.GetCell(ctx, id)
	if err != nil {
		return nil, err
	}

	cell.Status = status

	req := &api.UpdateCellRequest{
		Identity:   id,
		Cell:       cell,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"status"}},
	}

	c.logger.Println("updating cell...")
	resp, err := (*connClient).UpdateCell(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
