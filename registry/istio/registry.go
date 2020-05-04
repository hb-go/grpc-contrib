package istio

import (
	"net"
	"strings"

	"github.com/hb-go/grpc-contrib/registry"
	"google.golang.org/grpc/resolver"
)

func init() {
	registry.DefaultRegistry = NewRegistry()
}

type istioRegistry struct {
	registry.MockRegistry
}

// NewBuilder return resolver builder
func (r *istioRegistry) NewBuilder() resolver.Builder {
	return newBuilder()
}

// NewTarget return grpc.Dial target
func (r *istioRegistry) NewTarget(s *registry.Service, opts ...registry.Option) string {
	options := registry.Options{}
	for _, o := range opts {
		o(&options)
	}

	addr := s.Name
	addr = strings.ToLower(addr)
	addr = strings.Replace(addr, ".", "-", -1)
	addr = strings.Replace(addr, "_", "-", -1)

	if len(options.Addrs) > 0 {
		if _, port, err := net.SplitHostPort(options.Addrs[0]); err == nil && port != "" {
			addr = addr + ":" + port
		}
	}

	return schema + ":///" + addr
}

func NewRegistry() registry.Registry {
	return &istioRegistry{}
}
