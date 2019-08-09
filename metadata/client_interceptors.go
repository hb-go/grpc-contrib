package metadata

import (
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryClientInterceptor returns a new unary client interceptor for metadata.
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateOptions(opts)
	return func(parentCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md := metautils.ExtractIncoming(parentCtx)
		kv := []string{}
		if len(o.prefixes) > 0 {
			for k, _ := range md {
				if _, ok := o.headers[k]; ok {
					if v := md.Get(k); v != "" {
						kv = append(kv, k, v)
					}
					continue
				}

				for prefix, _ := range o.prefixes {
					if strings.HasPrefix(k, prefix) {
						if v := md.Get(k); v != "" {
							kv = append(kv, k, v)
						}
						break
					}
				}

			}
		} else {
			for k, _ := range o.headers {
				if v := md.Get(k); v != "" {
					kv = append(kv, k, v)
				}
			}
		}

		if len(kv) == 0 {
			return invoker(parentCtx, method, req, reply, cc, opts...)
		}

		newCtx := metadata.AppendToOutgoingContext(parentCtx, kv...)
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor returns a new streaming client interceptor for metadata.
func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateOptions(opts)
	return func(parentCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		md := metautils.ExtractIncoming(parentCtx)
		kv := []string{}
		if len(o.prefixes) > 0 {
			for k, _ := range md {
				if _, ok := o.headers[k]; ok {
					if v := md.Get(k); v != "" {
						kv = append(kv, k, v)
					}
					continue
				}

				for prefix, _ := range o.prefixes {
					if strings.HasPrefix(k, prefix) {
						if v := md.Get(k); v != "" {
							kv = append(kv, k, v)
						}
						break
					}
				}

			}
		} else {
			for k, _ := range o.headers {
				if v := md.Get(k); v != "" {
					kv = append(kv, k, v)
				}
			}
		}

		if len(kv) == 0 {
			return streamer(parentCtx, desc, cc, method, opts...)
		}

		newCtx := metadata.AppendToOutgoingContext(parentCtx, kv...)
		return streamer(newCtx, desc, cc, method, opts...)
	}
}
