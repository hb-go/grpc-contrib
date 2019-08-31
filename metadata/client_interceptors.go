package metadata

import (
	"context"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryClientInterceptor returns a new unary client interceptor for metadata.
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateOptions(opts)
	return func(parentCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		kv := metadataFromContext(o, parentCtx)
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
		kv := metadataFromContext(o, parentCtx)
		if len(kv) == 0 {
			return streamer(parentCtx, desc, cc, method, opts...)
		}

		newCtx := metadata.AppendToOutgoingContext(parentCtx, kv...)
		return streamer(newCtx, desc, cc, method, opts...)
	}
}

func metadataFromContext(o *options, ctx context.Context) []string {
	md := metautils.ExtractIncoming(ctx)
	kv := []string{}
	if len(o.prefixes) > 0 {
		for k, _ := range md {
			v := md.Get(k)
			if v == "" {
				continue
			}

			if newKey, ok := o.headers[k]; ok {
				if newKey == "" {
					kv = append(kv, k, v)
				} else {
					kv = append(kv, newKey, v)
				}
				continue
			}

			for prefix, newPrefix := range o.prefixes {
				if strings.HasPrefix(k, prefix) {
					newKey := k
					if newPrefix != prefix {
						newKey = strings.Replace(k, prefix, newPrefix, 1)
					}
					kv = append(kv, newKey, v)
					break
				}
			}

		}
	} else {
		for k, newKey := range o.headers {
			if v := md.Get(k); v != "" {
				if newKey == "" {
					kv = append(kv, k, v)
				} else {
					kv = append(kv, newKey, v)
				}
			}
		}
	}

	return kv
}
