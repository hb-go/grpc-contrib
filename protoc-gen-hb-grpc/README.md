# protoc-gen-hb-grpc

```bash
# install
go get -u github.com/hb-go/grpc-contrib/protoc-gen-hb-grpc
```

```bash
# Registry
protoc --proto_path=.:$GOPATH/src --hb-grpc_out=plugins=desc+registry:. proto/service.proto

# 自定义后缀名，默认.pb.grpc.hb.go
protoc --proto_path=.:$GOPATH/src --hb-grpc_out=plugins=registry,suffix=.pb.hb.grpc.go:. proto/service.proto
```
