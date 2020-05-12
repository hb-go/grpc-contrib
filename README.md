# gRPC Contrib

## Features

Path | 功能 | 说明
----|----|----
[client](client) | client 连接池 | -
[log](log) | logger | `zap`
[metadata](metadata) | 元数据转换与传递 | 支持 `client` 和 `gateway`插件
[protoc-gen-hb-grpc](protoc-gen-hb-grpc) | 代码生成 | 1.私有 service desc 导出<br/> 2.Registry 服务发现相关代码生成
[registry](registry) | 服务发现 | 支持 etcd、consul，以及 [istio](https://github.com/istio/istio) 环境
  
## Gen proto
```bash
protoc --proto_path=.:$GOPATH/src --go_out=plugins=grpc:. proto/service.proto
```

### 使用`protoc-gen-hb-grpc`
```bash
# install
go get -u github.com/hb-go/grpc-contrib/protoc-gen-hb-grpc
```

```bash
# 导出ServiceDesc & Registry
protoc --proto_path=.:$GOPATH/src --go_out=plugins=grpc:. \
--hb-grpc_out=plugins=desc+registry:. \
proto/service.proto

# 自定义后缀名，默认.hb.grpc.go
protoc --proto_path=.:$GOPATH/src --hb-grpc_out=plugins=registry,suffix=.hb.grpc.go:. proto/service.proto
```

#### Proto
```shell script
protoc --proto_path=.:$GOPATH/src \
--go_out=plugins=grpc:. \
--hb-grpc_out=plugins=registry:. \
proto/service.proto
```
