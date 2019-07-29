package istio

import (
	"github.com/hb-go/grpc-contrib/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"net"
)

type istioRegistry struct {
}

func NewRegistry() registry.Registry {
	return &istioRegistry{}
}

// NewBuilder return resolver builder
func (r *istioRegistry) NewBuilder() resolver.Builder {
	return newBuilder()
}

// NewTarget return grpc.Dial target
func (r *istioRegistry) NewTarget(sd *grpc.ServiceDesc, opts ...registry.Option) string {
	options := registry.Options{}
	for _, o := range opts {
		o(&options)
	}

	addr := ""
	if len(options.Addr) > 0 && options.Addr[:1] == ":" {
		addr = options.Addr
	} else if _, port, err := net.SplitHostPort(options.Addr); err == nil {
		addr = ":" + port
	}

	return schema + ":///" + sd.ServiceName + addr
}

// Register
func (r *istioRegistry) Register(sd *grpc.ServiceDesc, opts ...registry.Option) error {
	return nil
}

// Deregister
func (r *istioRegistry) Deregister(sd *grpc.ServiceDesc, opts ...registry.Option) {
}
