package istio

import (
	"net"
	"strings"

	"github.com/hb-go/grpc-contrib/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

func init() {
	registry.DefaultRegistry = NewRegistry()
}

type istioRegistry struct {
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

	addr := sd.ServiceName
	addr = strings.ToLower(addr)
	addr = strings.Replace(addr, ".", "-", -1)
	addr = strings.Replace(addr, "_", "-", -1)

	if _, port, err := net.SplitHostPort(options.Addr); err == nil && port != "" {
		addr = addr + ":" + port
	}

	return schema + ":///" + addr
}

// Register
func (r *istioRegistry) Register(sd *grpc.ServiceDesc, opts ...registry.Option) error {
	return nil
}

// Deregister
func (r *istioRegistry) Deregister(sd *grpc.ServiceDesc, opts ...registry.Option) {
}

func NewRegistry() registry.Registry {
	return &istioRegistry{}
}
