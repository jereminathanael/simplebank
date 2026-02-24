package gapi

import (
	"context"

	"google.golang.org/grpc"
)

func GrpcLogger(
	ctx context.Context, 
	req any, info *grpc.UnaryServerInfo, 
	handler grpc.UnaryHandler,
) (resp any, err error) {
	return
}