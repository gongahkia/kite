package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gongahkia/kite/internal/observability"
)

// loggingInterceptor logs all unary RPC requests
func loggingInterceptor(logger *observability.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the request
		duration := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			statusCode = status.Code(err)
		}

		logger.WithFields(map[string]interface{}{
			"method":   info.FullMethod,
			"duration": duration.Milliseconds(),
			"status":   statusCode.String(),
		}).Infof("gRPC %s %dms", info.FullMethod, duration.Milliseconds())

		return resp, err
	}
}

// recoveryInterceptor recovers from panics in unary RPCs
func recoveryInterceptor(logger *observability.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.WithFields(map[string]interface{}{
					"method": info.FullMethod,
					"panic":  r,
				}).Error("Panic recovered in gRPC handler")
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// metricsInterceptor records metrics for unary RPCs
func metricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			statusCode = status.Code(err)
		}

		// Record metrics (placeholder - integrate with observability.Metrics)
		_ = duration
		_ = statusCode

		return resp, err
	}
}

// streamLoggingInterceptor logs all streaming RPC requests
func streamLoggingInterceptor(logger *observability.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		err := handler(srv, ss)

		duration := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			statusCode = status.Code(err)
		}

		logger.WithFields(map[string]interface{}{
			"method":   info.FullMethod,
			"duration": duration.Milliseconds(),
			"status":   statusCode.String(),
			"stream":   true,
		}).Infof("gRPC Stream %s %dms", info.FullMethod, duration.Milliseconds())

		return err
	}
}

// streamRecoveryInterceptor recovers from panics in streaming RPCs
func streamRecoveryInterceptor(logger *observability.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.WithFields(map[string]interface{}{
					"method": info.FullMethod,
					"panic":  r,
					"stream": true,
				}).Error("Panic recovered in gRPC stream handler")
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(srv, ss)
	}
}
