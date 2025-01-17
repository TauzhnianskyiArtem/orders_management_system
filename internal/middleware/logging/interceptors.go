package logging

import (
	"context"

	"github.com/moguchev/microservices_courcse/orders_management_system/pkg/logger"
	"google.golang.org/grpc"
)

func LogErrorUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		logCtx := logger.ToContext(context.Background(),
			logger.FromContext(ctx).With(
				"operation", info.FullMethod,
				"component", "middleware",
			),
		)

		logger.Debug(logCtx, "receive request")
		resp, err = handler(ctx, req)
		logger.Debug(logCtx, "handle request")

		if err != nil {
			// 4ХХ -> warn
			// 5ХХ -> Error
			logger.Error(logCtx, err.Error())
		}

		return
	}
}
