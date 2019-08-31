package metadata

import (
	"context"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/metadata"
)

func GatewayHeaderMatcherFunc(withDefault bool, opts ...Option) runtime.HeaderMatcherFunc {
	o := evaluateOptions(opts)
	return func(key string) (string, bool) {
		if withDefault {
			if k, ok := runtime.DefaultHeaderMatcher(key); ok {
				return k, ok
			}
		}

		if newKey, ok := o.headers[key]; ok {
			if newKey == "" {
				return key, ok
			} else {
				return newKey, ok
			}
		}

		for prefix, newPrefix := range o.prefixes {
			if strings.HasPrefix(key, prefix) {
				newKey := key
				if newPrefix != prefix {
					newKey = strings.Replace(key, prefix, newPrefix, 1)
				}

				return newKey, true
			}
		}

		return "", false

	}
}

func GatewayMetadataAnnotator(opts ...Option) func(context.Context, *http.Request) metadata.MD {
	o := evaluateOptions(opts)
	return func(ctx context.Context, req *http.Request) metadata.MD {
		kv := []string{}
		if len(o.prefixes) > 0 {
			for k, _ := range req.Header {
				v := req.Header.Get(k)
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
				if v := req.Header.Get(k); v != "" {
					if newKey == "" {
						kv = append(kv, k, v)
					} else {
						kv = append(kv, newKey, v)
					}
				}
			}
		}

		return metadata.Pairs(kv...)
	}
}
