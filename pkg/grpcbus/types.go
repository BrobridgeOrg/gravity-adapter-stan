package grpcbus

import "google.golang.org/grpc"

type GRPCPool interface {
	Get() (*grpc.ClientConn, error)
	Put(*grpc.ClientConn) error
}
