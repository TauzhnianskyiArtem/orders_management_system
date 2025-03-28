package server

import (
	"context"
	"fmt"
	"github.com/moguchev/microservices_courcse/orders_management_system/pkg/closer"
	"net"
	"net/http"
	"time"

	"github.com/bufbuild/protovalidate-go"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/moguchev/microservices_courcse/orders_management_system/internal/app/usecases/orders_management_system"
	pb "github.com/moguchev/microservices_courcse/orders_management_system/pkg/api/orders_management_system"
	"github.com/moguchev/microservices_courcse/orders_management_system/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	GRPCPort        string
	GRPCGatewayPort string

	ChainUnaryInterceptors []grpc.UnaryServerInterceptor
	UnaryInterceptors      []grpc.UnaryServerInterceptor
}

type Deps struct {
	OMSUsecase orders_management_system.UsecaseInterface
}

type Server struct {
	pb.UnimplementedOrdersManagementSystemServiceServer
	Deps

	validator *protovalidate.Validator

	grpc struct {
		lis    net.Listener
		server *grpc.Server
	}

	grpcGateway struct {
		lis    net.Listener
		server *http.Server
	}
}

func New(ctx context.Context, cfg Config, d Deps) (*Server, error) {
	srv := &Server{
		Deps: d,
	}

	{
		validator, err := protovalidate.New(
			protovalidate.WithDisableLazy(true),
			protovalidate.WithMessages(
				&pb.CreateOrderRequest{},
			),
		)
		if err != nil {
			return nil, fmt.Errorf("server: failed to initialize validator: %w", err)
		}
		srv.validator = validator
	}

	{
		grpcServerOptions := unaryInterceptorsToGrpcServerOptions(cfg.UnaryInterceptors...)
		grpcServerOptions = append(grpcServerOptions,
			grpc.ChainUnaryInterceptor(cfg.ChainUnaryInterceptors...),
		)

		grpcServer := grpc.NewServer(grpcServerOptions...)
		pb.RegisterOrdersManagementSystemServiceServer(grpcServer, srv)

		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", cfg.GRPCPort)
		if err != nil {
			return nil, fmt.Errorf("server: failed to listen: %v", err)
		}

		srv.grpc.server = grpcServer
		srv.grpc.lis = lis
	}

	{
		mux := runtime.NewServeMux()
		if err := pb.RegisterOrdersManagementSystemServiceHandlerServer(ctx, mux, srv); err != nil {
			return nil, fmt.Errorf("server: failed to register handler: %v", err)
		}

		httpServer := &http.Server{Handler: mux}

		lis, err := net.Listen("tcp", cfg.GRPCGatewayPort)
		if err != nil {
			return nil, fmt.Errorf("server: failed to listen: %v", err)
		}

		srv.grpcGateway.server = httpServer
		srv.grpcGateway.lis = lis
	}

	return srv, nil
}

func (s *Server) Run(ctx context.Context) error {

	go func() {
		closer.Add(func(ctx context.Context) error {
			done := make(chan struct{})
			go func() {
				s.grpc.server.GracefulStop()
				close(done)
			}()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-done:
				return nil
			}
		})
		logger.Info(ctx, "start serve", s.grpc.lis.Addr())
		if err := s.grpc.server.Serve(s.grpc.lis); err != nil {
			logger.Error(ctx, "server: serve grpc: %v", err)
		}
	}()

	go func() {
		closer.Add(s.grpcGateway.server.Shutdown)
		logger.Info(ctx, "start serve", s.grpcGateway.lis.Addr())
		if err := s.grpcGateway.server.Serve(s.grpcGateway.lis); err != nil {
			logger.Error(ctx, "server: serve grpc gateway: %v", err)
		}
	}()

	<-ctx.Done()

	logger.Info(ctx, "server: shutting down server gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := closer.CloseAll(shutdownCtx); err != nil {
		return fmt.Errorf("closer: %v", err)
	}

	logger.Info(ctx, "server: shutdown")

	return nil
}

func unaryInterceptorsToGrpcServerOptions(interceptors ...grpc.UnaryServerInterceptor) []grpc.ServerOption {
	opts := make([]grpc.ServerOption, 0, len(interceptors))
	for _, interceptor := range interceptors {
		opts = append(opts, grpc.UnaryInterceptor(interceptor))
	}
	return opts
}
