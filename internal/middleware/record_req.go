package middleware

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	"log"

	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"
	"io"
)

// 记录grpc请求
func UnaryServerInterceptor(writer io.Writer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		//if limiter.Limit() {
		//	return nil, status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_ratelimit middleware, please retry later.", info.FullMethod)
		//}

		bt, _ := json.Marshal(req)
		log.Println("FullMethod", info.FullMethod, string(bt))
		//log.Println("FullMethod", info.Server)
		return handler(ctx, req)
	}
}

func StreamServerInterceptor(writer io.Writer) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		//if limiter.Limit() {
		//	return status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_ratelimit middleware, please retry later.", info.FullMethod)
		//}
		//bt, _ := json.Marshal(stream.)
		//log.Println("FullMethod", info.FullMethod, string(bt))
		return handler(srv, stream)
	}
}
