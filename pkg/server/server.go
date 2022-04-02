package server

import (
	"demoserver/api"
	"demoserver/testdata"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc/reflection"

	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	ctx            context.Context
	serverAddr     string
	serverPort     string
	logger         *log.Logger
	serverClosedCh *chan struct{}

	api.UnimplementedOrchestratorServiceServer
}

func New(serverAddr, serverPort string, logger *log.Logger, apiClosedCh *chan struct{}) (s *Server) {
	return &Server{
		ctx:            context.TODO(),
		serverAddr:     serverAddr,
		serverPort:     serverPort,
		logger:         logger,
		serverClosedCh: apiClosedCh}
}

// Serve - Starts gRPC API server running on local host
// Blocks until context cancelled or fatal error
func (s *Server) Serve(ctx context.Context) {
	s.ctx = ctx

	// Set the Server's address and port
	address := fmt.Sprintf("%s:%s", s.serverAddr, s.serverPort)
	log.Default()
	s.logger.Printf("starting gRPC server on %s:%s", s.serverAddr, s.serverPort)

	// Start server
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		s.logger.Printf("failed to setup listener on tcp %s", address)
		return
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	api.RegisterOrchestratorServiceServer(grpcServer, s)

	err = grpcServer.Serve(listener)
	if err != nil {
		s.logger.Printf("grpc api-server stopped: %s", err.Error())
		close(*s.serverClosedCh)
	}

}

// func (s *Server) Status(ctx context.Context) (cell *api.Cell, err error) {
// 	s.logger.Printf("GetCell called with id %s", id.Uuid)
// }

func (s *Server) GetCell(ctx context.Context, id *api.IdMessage) (cell *api.Cell, err error) {
	s.logger.Printf("GetCell called with id %s", id.Uuid)
	cellID := testdata.TestCell.Identity.Identity
	if id.Uuid != cellID {
		return nil, fmt.Errorf("identity %s does not match expected %s", id.Uuid, cellID)
	}
	return &testdata.TestCell, nil
}

func (s *Server) HTTPReverseProxy() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	grpcServerEndpoint := fmt.Sprintf("%s:%s", s.serverAddr, s.serverPort)

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := api.RegisterOrchestratorServiceHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(":8081", mux)
}
