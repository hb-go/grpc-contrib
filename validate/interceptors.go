package validate

import (
	"context"
	"reflect"

	"google.golang.org/grpc"
)

// https://github.com/envoyproxy/protoc-gen-validate
// Client和Server的拦截验证

// UnaryClientInterceptor returns a new unary client interceptor for validate.
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	o := newOptions(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		t := reflect.TypeOf(req)
		if m, ok := t.MethodByName(o.funcName); ok {
			if e := m.Func.Call([]reflect.Value{reflect.ValueOf(req)}); len(e) > 0 {
				errInter := e[0].Interface()
				if errInter != nil {
					return errInter.(error)
				}
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// UnaryServerInterceptor returns a new unary server interceptor for validate.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := newOptions(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		t := reflect.TypeOf(req)
		if m, ok := t.MethodByName(o.funcName); ok {
			if e := m.Func.Call([]reflect.Value{reflect.ValueOf(req)}); len(e) > 0 {
				errInter := e[0].Interface()
				if errInter != nil {
					return nil, errInter.(error)
				}
			}
		}

		return handler(ctx, req)
	}
}
