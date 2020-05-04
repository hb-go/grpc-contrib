package client

import (
	"io"
	"time"

	"github.com/hb-go/grpc-contrib/registry"
	"google.golang.org/grpc"
)

var (
	pool *Pool
)

func init() {
	pool = NewPool(100, time.Second*30)
}

func SetPoolSize(size int) {
	pool.size = size
}

func SetPoolTTL(ttl time.Duration) {
	pool.ttl = int64(ttl.Seconds())
}

func Client(s *registry.Service, options ...Option) (*grpc.ClientConn, io.Closer, error) {
	opts := newOptions(options...)
	if len(opts.Name) > 0 {
		s.Name = opts.Name
	}

	addr := registry.NewTarget(s, opts.RegistryOptions...)

	conn, err := pool.Get(addr, opts.DialOptions...)
	if err != nil {
		return nil, nil, err
	}

	c := &funcCloser{
		CloseFunc: func() error {
			pool.Put(addr, conn, err)
			return nil
		},
	}

	return conn.GetCC(), c, nil
}

type funcCloser struct {
	CloseFunc func() error
}

func (c *funcCloser) Close() error {
	return c.CloseFunc()
}
