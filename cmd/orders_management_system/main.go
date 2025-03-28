package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_opentracing "github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/moguchev/microservices_courcse/orders_management_system/internal/app/repository/orders_storage"
	"github.com/moguchev/microservices_courcse/orders_management_system/internal/app/server"
	"github.com/moguchev/microservices_courcse/orders_management_system/internal/app/services/warehouses_management_system"
	"github.com/moguchev/microservices_courcse/orders_management_system/internal/app/usecases/orders_management_system"
	middleware_errors "github.com/moguchev/microservices_courcse/orders_management_system/internal/middleware/errors"
	middleware_logging "github.com/moguchev/microservices_courcse/orders_management_system/internal/middleware/logging"
	middleware_recovery "github.com/moguchev/microservices_courcse/orders_management_system/internal/middleware/recovery"
	middleware_tracing "github.com/moguchev/microservices_courcse/orders_management_system/internal/middleware/tracing"
	"github.com/moguchev/microservices_courcse/orders_management_system/pkg/logger"
	"github.com/moguchev/microservices_courcse/orders_management_system/pkg/postgres"
	jaeger_tracing "github.com/moguchev/microservices_courcse/orders_management_system/pkg/tracing"
	"github.com/moguchev/microservices_courcse/orders_management_system/pkg/transaction_manager"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	logger.SetLevel(zapcore.InfoLevel)

	logger.Info(ctx, "start app init")
	if err := jaeger_tracing.Init("orders-management-system"); err != nil {
		logger.Fatal(ctx, err)
	}

	dsn := os.Getenv("DB_DSN")

	pool, err := postgres.NewConnectionPool(ctx, dsn,
		postgres.WithMaxConnIdleTime(5*time.Minute),
		postgres.WithMaxConnLifeTime(time.Hour),
		postgres.WithMaxConnectionsCount(10),
		postgres.WithMinConnectionsCount(5),
	)
	if err != nil {
		logger.FatalKV(ctx, "can't connect to database", "error", err.Error(), "dsn", dsn)
	}

	txManager := transaction_manager.New(pool)

	storage := orders_storage.New(txManager)

	wmsClient := warehouses_management_system.NewClient()

	omsUsecase := orders_management_system.NewUsecase(orders_management_system.Deps{ // Dependency injection
		WarehouseManagementSystem: wmsClient,
		OrdersStorage:             storage,
		TransactionManager:        txManager,
	})

	config := server.Config{
		GRPCPort:        os.Getenv("GRPC_PORT"),
		GRPCGatewayPort: os.Getenv("HTTP_PORT"),
		ChainUnaryInterceptors: []grpc.UnaryServerInterceptor{
			grpc_opentracing.OpenTracingServerInterceptor(opentracing.GlobalTracer(), grpc_opentracing.LogPayloads()),
			middleware_logging.LogErrorUnaryInterceptor(),
			middleware_tracing.DebugOpenTracingUnaryServerInterceptor(true, true),
			middleware_recovery.RecoverUnaryInterceptor(),
		},
		UnaryInterceptors: []grpc.UnaryServerInterceptor{
			middleware_errors.ErrorsUnaryInterceptor(),
		},
	}

	srv, err := server.New(ctx, config, server.Deps{
		OMSUsecase: omsUsecase,
	})
	if err != nil {
		logger.Fatalf(ctx, "failed to create server: %v", err)
	}

	if err = srv.Run(ctx); err != nil {
		logger.Errorf(ctx, "run: %v", err)
	}
}
