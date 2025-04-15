OPENAPI_SPEC = docs/swagger.yaml
GEN_OUTPUT = internal/delivery/http/httpdto/api.gen.go

PROTOC_SPEC = docs/pvz.proto
PROTOC_OUT_DIR = internal/delivery/grpc/pvz_v1
PROTOC_GO_OUT = $(PROTOC_OUT_DIR)/pvz.pb.go
PROTOC_GRPC_OUT = $(PROTOC_OUT_DIR)/pvz_grpc.pb.go

.PHONY: generate lint lint-fix gen-install test unit-tests integration-tests deploy

gen-install:
	apt update && apt install -y protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
	go install go.uber.org/mock/mockgen@latest


generate-oapi:
	@oapi-codegen \
    		-generate types \
    		-package httpdto \
    		-o $(GEN_OUTPUT) \
    		$(OPENAPI_SPEC)

generate-proto:
	@protoc --proto_path=docs \
		--go_out=$(PROTOC_OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTOC_OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTOC_SPEC)

generate-mocks:
	@mockgen -destination=internal/service/mocks/auth_mock.go -source=internal/service/auth.go
	@mockgen -destination=internal/service/mocks/product_mock.go -source=internal/service/product.go
	@mockgen -destination=internal/service/mocks/pvz_mock.go -source=internal/service/pvz.go
	@mockgen -destination=internal/service/mocks/reception_mock.go -source=internal/service/reception.go

	@mockgen -destination=internal/domain/repositories/mocks/product_repo_mock.go -source=internal/domain/repositories/product_repo.go
	@mockgen -destination=internal/domain/repositories/mocks/pvz_repo_mock.go -source=internal/domain/repositories/pvz_repo.go
	@mockgen -destination=internal/domain/repositories/mocks/reception_repo_mock.go -source=internal/domain/repositories/reception_repo.go
	@mockgen -destination=internal/domain/repositories/mocks/user_repo_mock.go -source=internal/domain/repositories/user_repo.go

	@mockgen -destination=internal/pkg/auth/mocks/manager_mock.go -source=internal/pkg/auth/manager.go

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

generate: generate-oapi generate-proto generate-mocks

deploy:
	docker compose up --build

unit-tests: generate
	@echo Running unit tests
	@go test ./... --cover --tags=unit

integration-tests: generate
	@echo Running integration tests
	@go test ./... --cover --tags=integration

test: unit-tests integration-tests
