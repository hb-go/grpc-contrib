package metadata

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

func GatewayMetadataAnnotator(opts ...Option) func(context.Context, *http.Request) metadata.MD {
	o := evaluateOptions(opts)
	return func(ctx context.Context, req *http.Request) metadata.MD {
		kv := []string{}
		if len(o.prefixes) > 0 {
			for k, _ := range req.Header {
				if _, ok := o.headers[k]; ok {
					if v := req.Header.Get(k); v != "" {
						kv = append(kv, k, v)
					}
					continue
				}

				for prefix, _ := range o.prefixes {
					if strings.HasPrefix(k, prefix) {
						if v := req.Header.Get(k); v != "" {
							kv = append(kv, k, v)
						}
						break
					}
				}

			}
		} else {
			for k, _ := range o.headers {
				if v := req.Header.Get(k); v != "" {
					kv = append(kv, k, v)
				}
			}
		}

		return metadata.Pairs(kv...)
	}
}
