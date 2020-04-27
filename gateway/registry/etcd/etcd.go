// Package etcd provides an etcd service registry
package etcd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hb-go/grpc-contrib/gateway/registry"
	hash "github.com/mitchellh/hashstructure"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	"go.uber.org/zap"
)

var (
	prefix = "/grpc-gateway/registry/"
)

type etcdRegistry struct {
	client  *clientv3.Client
	options registry.Options

	sync.RWMutex
	register map[string]uint64
	leases   map[string]clientv3.LeaseID
}

func NewRegistry(opts ...registry.Option) registry.Registry {
	e := &etcdRegistry{
		options:  registry.Options{},
		register: make(map[string]uint64),
		leases:   make(map[string]clientv3.LeaseID),
	}
	configure(e, opts...)
	return e
}

func configure(e *etcdRegistry, opts ...registry.Option) error {
	config := clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	}

	for _, o := range opts {
		o(&e.options)
	}

	if e.options.Timeout == 0 {
		e.options.Timeout = 5 * time.Second
	}

	if e.options.Secure || e.options.TLSConfig != nil {
		tlsConfig := e.options.TLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		config.TLS = tlsConfig
	}

	if e.options.Context != nil {
		u, ok := e.options.Context.Value(authKey{}).(*authCreds)
		if ok {
			config.Username = u.Username
			config.Password = u.Password
		}
		cfg, ok := e.options.Context.Value(logConfigKey{}).(*zap.Config)
		if ok && cfg != nil {
			config.LogConfig = cfg
		}
	}

	var cAddrs []string

	for _, address := range e.options.Addrs {
		if len(address) == 0 {
			continue
		}
		addr, port, err := net.SplitHostPort(address)
		if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
			port = "2379"
			addr = address
			cAddrs = append(cAddrs, net.JoinHostPort(addr, port))
		} else if err == nil {
			cAddrs = append(cAddrs, net.JoinHostPort(addr, port))
		}
	}

	// if we got addrs then we'll update
	if len(cAddrs) > 0 {
		config.Endpoints = cAddrs
	}

	cli, err := clientv3.New(config)
	if err != nil {
		return err
	}
	e.client = cli
	return nil
}

func encode(s *registry.Service) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decode(ds []byte) *registry.Service {
	var s *registry.Service
	json.Unmarshal(ds, &s)
	return s
}

func versionPath(s, v string) string {
	service := strings.Replace(s, "/", "-", -1)
	version := strings.Replace(v, "/", "-", -1)
	return path.Join(prefix, service, version)
}

func servicePath(s string) string {
	return path.Join(prefix, strings.Replace(s, "/", "-", -1))
}

func (e *etcdRegistry) Init(opts ...registry.Option) error {
	return configure(e, opts...)
}

func (e *etcdRegistry) Options() registry.Options {
	return e.options
}

func (e *etcdRegistry) registerService(s *registry.Service, opts ...registry.RegisterOption) error {
	if len(s.Methods) == 0 {
		return errors.New("require at least one method")
	}

	// check existing lease cache
	e.RLock()
	leaseID, ok := e.leases[s.Name+s.Version]
	e.RUnlock()

	if !ok {
		// missing lease, check if the key exists
		ctx, cancel := context.WithTimeout(context.Background(), e.options.Timeout)
		defer cancel()

		// look for the existing key
		rsp, err := e.client.Get(ctx, versionPath(s.Name, s.Version), clientv3.WithSerializable())
		if err != nil {
			return err
		}

		// get the existing lease
		for _, kv := range rsp.Kvs {
			if kv.Lease > 0 {
				leaseID = clientv3.LeaseID(kv.Lease)

				// decode the existing node
				srv := decode(kv.Value)
				if srv == nil || len(srv.Methods) == 0 {
					continue
				}

				// create hash of service; uint64
				h, err := hash.Hash(srv.Methods, nil)
				if err != nil {
					continue
				}

				// save the info
				e.Lock()
				e.leases[s.Name+s.Version] = leaseID
				e.register[s.Name+s.Version] = h
				e.Unlock()

				break
			}
		}
	}

	var leaseNotFound bool

	// renew the lease if it exists
	if leaseID > 0 {
		// if logger.V(logger.TraceLevel, logger.DefaultLogger) {
		// 	logger.Tracef("Renewing existing lease for %s %d", s.Name, leaseID)
		// }
		if _, err := e.client.KeepAliveOnce(context.TODO(), leaseID); err != nil {
			if err != rpctypes.ErrLeaseNotFound {
				return err
			}

			// if logger.V(logger.TraceLevel, logger.DefaultLogger) {
			// 	logger.Tracef("Lease not found for %s %d", s.Name, leaseID)
			// }
			// lease not found do register
			leaseNotFound = true
		}
	}

	// create hash of service; uint64
	h, err := hash.Hash(s.Methods, nil)
	if err != nil {
		return err
	}

	// get existing hash for the service node
	e.Lock()
	v, ok := e.register[s.Name+s.Version]
	e.Unlock()

	// the service is unchanged, skip registering
	if ok && v == h && !leaseNotFound {
		// if logger.V(logger.TraceLevel, logger.DefaultLogger) {
		// 	logger.Tracef("Service %s node %s unchanged skipping registration", s.Name, node.Id)
		// }
		return nil
	}

	var options registry.RegisterOptions
	for _, o := range opts {
		o(&options)
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.options.Timeout)
	defer cancel()

	var lgr *clientv3.LeaseGrantResponse
	if options.TTL.Seconds() > 0 {
		// get a lease used to expire keys since we have a ttl
		lgr, err = e.client.Grant(ctx, int64(options.TTL.Seconds()))
		if err != nil {
			return err
		}
	}

	// if logger.V(logger.TraceLevel, logger.DefaultLogger) {
	// 	logger.Tracef("Registering %s id %s with lease %v and leaseID %v and ttl %v", service.Name, node.Id, lgr, lgr.ID, options.TTL)
	// }
	// create an entry for the node
	if lgr != nil {
		_, err = e.client.Put(ctx, versionPath(s.Name, s.Version), encode(s), clientv3.WithLease(lgr.ID))
	} else {
		_, err = e.client.Put(ctx, versionPath(s.Name, s.Version), encode(s))
	}
	if err != nil {
		return err
	}

	e.Lock()
	// save our hash of the service
	e.register[s.Name+s.Version] = h
	// save our leaseID of the service
	if lgr != nil {
		e.leases[s.Name+s.Version] = lgr.ID
	}
	e.Unlock()

	return nil
}

func (e *etcdRegistry) Deregister(s *registry.Service) error {
	e.Lock()
	// delete our hash of the service
	delete(e.register, s.Name+s.Version)
	// delete our lease of the service
	delete(e.leases, s.Name+s.Version)
	e.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), e.options.Timeout)
	defer cancel()

	// if logger.V(logger.TraceLevel, logger.DefaultLogger) {
	// 	logger.Tracef("Deregistering %s id %s", s.Name, node.Id)
	// }
	_, err := e.client.Delete(ctx, versionPath(s.Name, s.Version))
	if err != nil {
		return err
	}

	return nil
}

func (e *etcdRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	if len(s.Methods) == 0 {
		return errors.New("require at least one method")
	}

	// register service
	gerr := e.registerService(s, opts...)

	return gerr
}

func (e *etcdRegistry) GetService(name string) ([]*registry.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.options.Timeout)
	defer cancel()

	rsp, err := e.client.Get(ctx, servicePath(name)+"/", clientv3.WithPrefix(), clientv3.WithSerializable())
	if err != nil {
		return nil, err
	}

	if len(rsp.Kvs) == 0 {
		return nil, registry.ErrNotFound
	}

	versions := map[string]*registry.Service{}

	for _, n := range rsp.Kvs {
		if sn := decode(n.Value); sn != nil {
			_, ok := versions[sn.Version]
			if !ok {
				versions[sn.Version] = sn
			}
		}
	}

	services := make([]*registry.Service, 0, len(versions))
	for _, service := range versions {
		services = append(services, service)
	}

	return services, nil
}

func (e *etcdRegistry) ListServices() ([]*registry.Service, error) {
	versions := make(map[string]*registry.Service)

	ctx, cancel := context.WithTimeout(context.Background(), e.options.Timeout)
	defer cancel()

	rsp, err := e.client.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithSerializable())
	if err != nil {
		return nil, err
	}

	if len(rsp.Kvs) == 0 {
		return []*registry.Service{}, nil
	}

	for _, n := range rsp.Kvs {
		sn := decode(n.Value)
		if sn == nil {
			continue
		}
		_, ok := versions[sn.Name+sn.Version]
		if !ok {
			versions[sn.Name+sn.Version] = sn
			continue
		}
	}

	services := make([]*registry.Service, 0, len(versions))
	for _, service := range versions {
		services = append(services, service)
	}

	// sort the services
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })

	return services, nil
}

func (e *etcdRegistry) Watch(opts ...registry.WatchOption) (registry.Watcher, error) {
	return newEtcdWatcher(e, e.options.Timeout, opts...)
}

func (e *etcdRegistry) String() string {
	return "etcd"
}
