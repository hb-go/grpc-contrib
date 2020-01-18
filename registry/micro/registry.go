package micro

import (
	"net"
	"strings"
	"sync"

	mregistry "github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/registry/etcd"
	"github.com/micro/go-micro/util/addr"
	mnet "github.com/micro/go-micro/util/net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"

	"github.com/hb-go/grpc-contrib/registry"
)

var (
	deregisterCh   = make(chan struct{})
	deregisterOnce = sync.Once{}
)

func init() {
	mregistry.DefaultRegistry = etcd.NewRegistry()
	registry.DefaultRegistry = NewRegistry()
}

type microRegistry struct {
	wg *sync.WaitGroup
}

// NewBuilder return resolver builder
func (r *microRegistry) NewBuilder() resolver.Builder {
	return newBuilder()
}

// NewTarget return grpc.Dial target
func (r *microRegistry) NewTarget(sd *grpc.ServiceDesc, opts ...registry.Option) string {
	options := registry.Options{}
	for _, o := range opts {
		o(&options)
	}

	if options.Version == "" {
		return schema + ":///" + sd.ServiceName
	}

	return schema + ":///" + sd.ServiceName + "?version=" + options.Version
}

// Register
func (r *microRegistry) Register(sd *grpc.ServiceDesc, opts ...registry.Option) error {
	options := registry.Options{}
	for _, o := range opts {
		o(&options)
	}

	var err error
	service, err := newMicroService(sd, options)
	if err != nil {
		return err
	}

	// TODO 重复注册

	// register service
	grpclog.Infof("register service: %v", service)
	err = mregistry.Register(service)
	if err != nil {
		return err
	}

	// wait deregister then delete
	r.wg.Add(1)
	go func(wg *sync.WaitGroup) {
		<-deregisterCh
		err := mregistry.Deregister(service)
		if err != nil {
			grpclog.Warningf("grpc-contrib.registry.go-micro: deregister error: %v", err)
		}

		wg.Done()
	}(r.wg)

	return nil
}

// Deregister
func (r *microRegistry) Deregister(sd *grpc.ServiceDesc, opts ...registry.Option) {
	options := registry.Options{}
	for _, o := range opts {
		o(&options)
	}

	if sd != nil {
		// 注销单个服务
		if service, err := newMicroService(sd, options); err != nil {
			grpclog.Warningf("grpc-contrib.registry.go-micro: deregister error: %v", err)
		} else {
			err := mregistry.Deregister(service)
			if err != nil {
				grpclog.Warningf("grpc-contrib.registry.go-micro: deregister error: %v", err)
			}
		}
	} else {
		// 注销全部服务
		deregisterOnce.Do(func() {
			close(deregisterCh)
		})

		r.wg.Wait()
	}
}

func newMicroService(sd *grpc.ServiceDesc, options registry.Options) (*mregistry.Service, error) {
	var err error
	var host, port string

	if cnt := strings.Count(options.Addr, ":"); cnt >= 1 {
		// ipv6 address in format [host]:port or ipv4 host:port
		host, port, err = net.SplitHostPort(options.Addr)
		if err != nil {
			return nil, err
		}
	} else {
		host = options.Addr
	}

	address, err := addr.Extract(host)
	if err != nil {
		return nil, err
	}
	address = mnet.HostPort(address, port)

	node := &mregistry.Node{
		Id:      sd.ServiceName + "-" + address,
		Address: address,
	}

	service := &mregistry.Service{
		Name:    sd.ServiceName,
		Version: options.Version,
		Nodes:   []*mregistry.Node{node},
	}

	return service, nil
}

func NewRegistry() registry.Registry {
	return &microRegistry{
		wg: &sync.WaitGroup{},
	}
}
