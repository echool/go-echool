package echool

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

// GrpcClient ...
type GrpcClient struct {
	target string
}

// NewClient ...
func NewClient(target string) *GrpcClient {
	return &GrpcClient{target: target}
}

// GetConnection ...
func (c *GrpcClient) GetConnection(options ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithBalancerName(roundrobin.Name)}
	opts = append(opts, options...)

	return grpc.DialContext(context.Background(), c.target, opts...)
}
