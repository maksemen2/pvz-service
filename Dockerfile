FROM golang:1.23

WORKDIR ${GOPATH}/pvz-service/
COPY . ${GOPATH}/pvz-service/

RUN go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest # Генерация кода из swagger.yaml

# Генерация кода из pvz.proto
RUN apt-get update && apt-get install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN oapi-codegen \
    -generate types \
    -package httpdto \
    -o internal/delivery/http/httpdto/api.gen.go \
    docs/swagger.yaml

RUN protoc --proto_path=docs \
    --go_out=internal/delivery/grpc/pvz_v1 \
    --go_opt=paths=source_relative \
    --go-grpc_out=internal/delivery/grpc/pvz_v1 \
    --go-grpc_opt=paths=source_relative \
    docs/pvz.proto


RUN go build -o /build ./cmd

EXPOSE 8080 3000 9000

CMD ["/build"]