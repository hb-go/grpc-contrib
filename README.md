# gRPC Contrib

## Gen proto
```bash
protoc -I/usr/local/include -I. \
  -I$GOPATH/src \
  --go_out=plugins=grpc:. \ 
  test/proto/*.proto
```