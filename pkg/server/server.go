package server

import (
	"demoserver/api"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	cells = make(map[string]*api.Cell, 0)
)

type Server struct {
	// ctx            context.Context
	serverAddr     string
	serverPort     string
	proxyPort      string
	logger         *log.Logger
	serverClosedCh *chan struct{}

	api.UnimplementedCellServiceServer
}

func New(serverAddr, serverPort, proxyPort string, logger *log.Logger, apiClosedCh *chan struct{}) (s *Server) {
	return &Server{
		// ctx:            context.TODO(),
		serverAddr:     serverAddr,
		serverPort:     serverPort,
		proxyPort:      proxyPort,
		logger:         logger,
		serverClosedCh: apiClosedCh}
}

// Serve - Starts gRPC API server running on local host
// Blocks until context cancelled or fatal error
func (s *Server) Serve(ctx context.Context) {

	// Set the Server's address and port
	address := fmt.Sprintf("%s:%s", s.serverAddr, s.serverPort)
	s.logger.Printf("starting gRPC server on %s", address)

	// Start server
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		s.logger.Printf("error: %s", address)
		return
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	api.RegisterCellServiceServer(grpcServer, s)

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
		if err := listener.Close(); err != nil {
			s.logger.Printf("failed to close %s: %v", address, err)
		}
		close(*s.serverClosedCh)
	}()

	err = grpcServer.Serve(listener)
	if err != nil {
		s.logger.Printf("gRPCs server stopped: %s", err.Error())
	}

}

// ServeReverseProxy - Starts the Reverse-proxy Server running on
// local host. Blocks until context cancelled or fatal error
func (s *Server) ServeReverseProxy(ctx context.Context) {

	// Set the Server's address and port
	grpcAddress := fmt.Sprintf("%s:%s", s.serverAddr, s.serverPort)
	address := fmt.Sprintf("%s:%s", s.serverAddr, s.proxyPort)
	s.logger.Printf("starting reverse-proxy server on %s", address)

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := api.RegisterCellServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		s.logger.Printf("error: %s", address)
	}

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		s.logger.Println("Shutting down the http gateway server")
		if err := server.Shutdown(context.Background()); err != nil {
			s.logger.Printf("failed to shutdown http gateway server: %v", err)
		}
		close(*s.serverClosedCh)
	}()

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		s.logger.Printf("failed to listen and serve: %v", err)
	}
}

// Status - Returns true if server is healthy with timestamp of access time
func (s *Server) Status(ctx context.Context, val *emptypb.Empty) (resp *api.ServerHealthResponse, err error) {
	t := time.Now()
	s.logger.Printf("Health check called at %s", t.Format("2006-01-02 15:04:05"))
	return &api.ServerHealthResponse{
		Healthy:     true,
		CurrentTime: timestamppb.New(t),
	}, nil
}

// CreateCell - Creates a Cell. Returns an error if the UUID is
func (s *Server) CreateCell(ctx context.Context, req *api.CreateCellRequest) (resp *api.Cell, err error) {

	cell := req.Cell
	if req.Cell.Identity == nil || req.Cell.Identity.Uuid == "" {
		// Assign identity
		cell.Identity.Uuid = uuid.New().String()
	}

	s.logger.Printf("Creating cells with id %s", cell.Identity.Uuid)
	cells[cell.Identity.Uuid] = cell

	return req.Cell, nil
}

// GetCell - Returns a Cell resource identified by the ID
func (s *Server) GetCell(ctx context.Context, req *api.GetCellRequest) (resp *api.Cell, err error) {
	s.logger.Printf("Get Cell called with id %s", req.Identity.Uuid)

	cell, ok := cells[req.Identity.Uuid]
	if !ok {
		return nil, fmt.Errorf("cell with ID %s not found", req.Identity.Uuid)
	}

	return cell, nil
}

// UpdateCell - Updates a Cell identified by the ID
func (s *Server) UpdateCell(ctx context.Context, req *api.UpdateCellRequest) (resp *api.Cell, err error) {
	s.logger.Printf("Update Cell called with id %s", req.Identity.Uuid)

	cell, ok := cells[req.Identity.Uuid]
	if !ok {
		return nil, fmt.Errorf("cell with ID %s not found", req.Identity.Uuid)
	}

	err = applyFieldMask(cell, req.Cell, req.UpdateMask)
	if err != nil {
		return nil, err
	}

	return cell, nil
}

type InvalidMaskPath struct {
	Path string
	Name protoreflect.Name
}

func (e *InvalidMaskPath) Error() string {
	return fmt.Sprintf("invalid mask path %s for object of type %s", e.Path, e.Name)
}

type InvalidMaskTypes struct {
	Patchee protoreflect.Name
	Patcher protoreflect.Name
}

func (e *InvalidMaskTypes) Error() string {
	return fmt.Sprintf("patchee %T and patcher %T must be same type", e.Patchee, e.Patcher)
}

func applyFieldMask(patchee, patcher proto.Message, mask *field_mask.FieldMask) (err error) {
	if mask == nil {
		return
	}
	if patchee.ProtoReflect().Descriptor().FullName() != patcher.ProtoReflect().Descriptor().FullName() {
		return &InvalidMaskTypes{
			Patchee: patchee.ProtoReflect().Descriptor().FullName().Name(),
			Patcher: patcher.ProtoReflect().Descriptor().FullName().Name(),
		}
	}

	for _, path := range mask.GetPaths() {
		patcherField, patcherParent, err := getField(patcher.ProtoReflect(), path)
		if err != nil {
			return err
		}
		patcheeField, patcheeParent, err := getField(patchee.ProtoReflect(), path)
		if err != nil {
			return err
		}
		patcheeParent.Set(patcheeField, patcherParent.Get(patcherField))
	}

	return nil
}

func getField(msg protoreflect.Message, path string) (field protoreflect.FieldDescriptor, parent protoreflect.Message, err error) {
	fields := msg.Descriptor().Fields()
	parent = msg
	names := strings.Split(path, ".")
	for i, name := range names {
		field = fields.ByName(protoreflect.Name(name))

		if i < len(names)-1 {
			parent = parent.Get(field).Message()
			fields = field.Message().Fields()
		}
	}

	if field == nil {
		return nil, nil, &InvalidMaskPath{Path: path, Name: parent.Descriptor().FullName().Name()}
	}

	return field, parent, nil
}
