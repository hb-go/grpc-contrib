package registry

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/time/rate"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
)

const schema = "registry"
const watchLimit = 1.0
const watchBurst = 3
const queryValSeq = "|"

// implementation of grpc.resolve.Builder
type registryBuilder struct {
	registry Registry

	mu        sync.RWMutex
	resolvers map[string]*service
}

type service struct {
	name   string
	target resolver.Target

	builder *registryBuilder

	mu       sync.RWMutex
	watching bool
	nodes    map[string][]resolver.Address

	conns     sync.Map
	connIndex int64
}

type registryResolver struct {
	service  *service
	versions []string

	index int64
	cc    resolver.ClientConn
}

// Scheme
func (b *registryBuilder) Scheme() string {
	return schema
}

// Build to resolver.Resolver
// target: {schema}://[authority]/{serviceName}[?version=v1]
// target使用query参数做version筛选
func (b *registryBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var serviceName string
	var serviceVersion []string
	if c := strings.Count(target.Endpoint, "?"); c > 0 {
		if u, err := url.Parse(target.Endpoint); err == nil {
			serviceName = u.Path
			query := u.Query()
			if v := query.Get("version"); len(v) > 0 {
				val := query.Get("version")
				serviceVersion = strings.Split(val, queryValSeq)
			}
		}
	} else {
		serviceName = target.Endpoint
	}

	// TODO resolver 需要根据 endpoint 记录 ClientConn，版本等筛选信息属于 client

	b.mu.Lock()
	s, ok := b.resolvers[serviceName]
	if ok {
		b.mu.Unlock()

		// 使用当前service nodes
		s.mu.Lock()
		var ccNodes []resolver.Address
		if len(serviceVersion) == 0 {
			ccNodes = make([]resolver.Address, 0, len(s.nodes))
			for _, v := range s.nodes {
				ccNodes = append(ccNodes, v...)
			}
		} else {
			for _, v := range serviceVersion {
				if nodes, ok := s.nodes[v]; ok {
					ccNodes = append(ccNodes, nodes...)
				}
			}
		}

		// TODO 检查watching状态?
		if !s.watching {
			err := s.watch()
			if err != nil {
				s.mu.Unlock()
				return nil, err
			}
			s.watching = true
		}
		s.mu.Unlock()

		cc.UpdateState(resolver.State{Addresses: ccNodes})
	} else {
		s = &service{
			name:    serviceName,
			target:  target,
			builder: b,
			nodes:   make(map[string][]resolver.Address),
		}
		b.resolvers[s.name] = s

		s.mu.Lock()
		b.mu.Unlock()

		// 从registry获取services
		services, err := b.registry.GetService(s.name)
		if err != nil {
			s.mu.Unlock()
			return nil, err
		}

		count := 0
		for _, svc := range services {
			nodes := make([]resolver.Address, 0, len(svc.Nodes))
			for _, n := range svc.Nodes {
				addr := resolver.Address{
					Addr: n.Address,
				}
				nodes = append(nodes, addr)
			}
			s.nodes[svc.Version] = nodes
			count++
		}

		var ccNodes []resolver.Address
		if len(serviceVersion) == 0 {
			ccNodes = make([]resolver.Address, 0, count)
			for _, v := range s.nodes {
				ccNodes = append(ccNodes, v...)
			}
		} else {
			for _, v := range serviceVersion {
				if nodes, ok := s.nodes[v]; ok {
					ccNodes = append(ccNodes, nodes...)
				}
			}
		}

		err = s.watch()
		if err != nil {
			s.mu.Unlock()
			return nil, err
		}

		s.watching = true
		s.mu.Unlock()

		cc.UpdateState(resolver.State{Addresses: ccNodes})
	}

	index := atomic.AddInt64(&s.connIndex, 1)
	r := &registryResolver{
		service:  s,
		versions: serviceVersion,
		cc:       cc,
		index:    index,
	}

	s.conns.Store(index, r)
	return r, nil
}

// ResolveNow
func (r *registryResolver) ResolveNow(rn resolver.ResolveNowOptions) {
}

// Close
func (r *registryResolver) Close() {
	r.service.conns.Delete(r.index)
}

func (s *service) watch() error {
	watcher, err := s.builder.registry.Watch(WatchService(s.name))
	if err != nil {
		return err
	}

	go func(watcher Watcher) {
		limiter := rate.NewLimiter(rate.Limit(watchLimit), watchBurst)
		for {
			limiter.Wait(context.Background())
			if result, err := watcher.Next(); err == nil {
				if err := s.process(result); err == nil {
					s.conns.Range(func(key, value interface{}) bool {
						if r, ok := value.(*registryResolver); ok {
							var allNodes []resolver.Address
							if len(r.versions) == 0 {
								for _, v := range s.nodes {
									allNodes = append(allNodes, v...)
								}
							} else {
								for _, v := range r.versions {
									if nodes, ok := s.nodes[v]; ok {
										allNodes = append(allNodes, nodes...)
									}
								}
							}

							r.cc.UpdateState(resolver.State{Addresses: allNodes})
						} else {
							grpclog.Warning("grpc-contrib.registry.go-micro: microResolver conv error")
						}

						return true
					})
				} else {
					grpclog.Warningf("grpc-contrib.registry.go-micro: %v", err)
				}
			} else {
				grpclog.Warningf("grpc-contrib.registry.go-micro: resolver watch error: %v", err)
				if err == ErrWatcherStopped {
					return
				}
			}
		}
	}(watcher)

	return nil
}

func (s *service) process(res *Result) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch res.Action {
	case "create", "update":
		nodes := make([]resolver.Address, 0, len(res.Service.Nodes))
		for _, n := range res.Service.Nodes {
			node := resolver.Address{
				Addr: n.Address,
			}
			nodes = append(nodes, node)
		}

		// append old nodes to new service
		if curNodes, ok := s.nodes[res.Service.Version]; ok {
			for _, node := range curNodes {
				var seen bool
				for _, n := range nodes {
					if node.Addr == n.Addr {
						seen = true
						break
					}
				}

				if !seen {
					nodes = append(nodes, node)
				}
			}
		}

		s.nodes[res.Service.Version] = nodes
		return nil

	case "delete":
		if curNodes, ok := s.nodes[res.Service.Version]; !ok {
			return nil
		} else {
			var nodes []resolver.Address

			// filter cur nodes to remove the dead one
			for _, cur := range curNodes {
				var seen bool
				for _, del := range res.Service.Nodes {
					if del.Address == cur.Addr {
						seen = true
						break
					}
				}
				if !seen {
					nodes = append(nodes, cur)
				}
			}

			s.nodes[res.Service.Version] = nodes
			return nil
		}
	default:
		return errors.New("un supported result action")
	}
}

// newBuilder return resolver builder
func newBuilder(r Registry) resolver.Builder {
	return &registryBuilder{
		registry:  r,
		resolvers: make(map[string]*service),
	}
}

func RegisterBuilder(r Registry) {
	b := newBuilder(r)
	resolver.Register(b)
}
