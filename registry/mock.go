package registry

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type mockRegistry struct {
}

// NewTarget return grpc.Dial target
func (r *mockRegistry) NewTarget(sd *grpc.ServiceDesc, opts ...Option) string {
	grpclog.Error("grpc-contrib.registry: default mock registry unrealized")
	return ""
}

// Register
func (r *mockRegistry) Register(sd *grpc.ServiceDesc, opts ...Option) error {
	grpclog.Error("grpc-contrib.registry: default mock registry unrealized")
	return nil
}

// Deregister
func (r *mockRegistry) Deregister(sd *grpc.ServiceDesc, opts ...Option) {
	grpclog.Error("grpc-contrib.registry: default mock registry unrealized")
}

func NewRegistry() Registry {
	return &mockRegistry{}
}
