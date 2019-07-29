package registry

import (
	"google.golang.org/grpc"
)

var (
	DefaultRegistry = NewRegistry()
)

type Registry interface {
	// 生成grpc.Dial target
	NewTarget(sd *grpc.ServiceDesc, opts ...Option) string

	// 注册服务
	Register(sd *grpc.ServiceDesc, opts ...Option) error
	// 注销服务
	// sd=nil，注销全部服务
	Deregister(sd *grpc.ServiceDesc, opts ...Option)
}

func NewTarget(sd *grpc.ServiceDesc, opts ...Option) string {
	return DefaultRegistry.NewTarget(sd, opts...)
}

func Register(sd *grpc.ServiceDesc, opts ...Option) error {
	return DefaultRegistry.Register(sd, opts...)
}

func Deregister(sd *grpc.ServiceDesc, opts ...Option) {
	DefaultRegistry.Deregister(sd, opts...)
}
