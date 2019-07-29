package istio

import (
	"errors"
	"net"

	"google.golang.org/grpc/resolver"
)

const schema = "istio"

func init() {
	r := newBuilder()
	resolver.Register(r)
}

// implementation of grpc.resolve.Builder
type istioBuilder struct {
}

// Scheme
func (b *istioBuilder) Scheme() string {
	return schema
}

// Build to resolver.Resolver
// target: {schema}://[authority]/{serviceName}:{port}
// target使用query参数做version筛选
func (b *istioBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	if host, port, err := net.SplitHostPort(target.Endpoint); err != nil {
		return nil, err
	} else if host == "" || port == "" {
		return nil, errors.New("host and port must be set. e.g. istio:///service-name:8080")
	}

	address := resolver.Address{Addr: target.Endpoint}
	cc.UpdateState(resolver.State{Addresses: []resolver.Address{address}})

	return nil, nil
}

// NewBuilder return resolver builder
func newBuilder() resolver.Builder {
	return &istioBuilder{}
}
