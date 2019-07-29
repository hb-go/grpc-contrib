# protoc-gen-hb-grpc

```bash
# install
go get -u github.com/hb-go/grpc-contrib/protoc-gen-hb-grpc
```

```bash
# Registry
protoc --proto_path=.:$GOPATH/src --hb-grpc_out=plugins=registry:. proto/service.proto

# 自定义后缀名，默认.hb.grpc.go
protoc --proto_path=.:$GOPATH/src --hb-grpc_out=plugins=registry,suffix=.hb.grpc.go:. proto/service.proto
```