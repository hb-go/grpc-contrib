package registry

import (
	"github.com/hb-go/grpc-contrib/registry/micro"
)

func NewRegistry() Registry {
	return micro.NewRegistry()
}
