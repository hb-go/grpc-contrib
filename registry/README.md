# gRPC服务发现

使用google.golang.org/grpc/resolver实现服务发现

- Feature
    - [x] [go-micro/registry](https://github.com/micro/go-micro/tree/master/registry)
        - [x] 服务发现
        - [x] 版本选择
    - [x] istio
    

## 使用
默认使用go-micro mdns注册服务

```go
// 服务注册与注销
registry.Register(&_ExampleService_serviceDesc, opts...)
registry.Deregister(&_ExampleService_serviceDesc, opts...)
registry.Deregister(nil)

// 服务发现
target := registry.NewTarget(&_ExampleService_serviceDesc, opts...)
conn, err := grpc.Dial(target, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name), grpc.WithBlock())
```