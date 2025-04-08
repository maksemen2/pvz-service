OPENAPI_SPEC = docs/swagger.yaml
GEN_OUTPUT = internal/delivery/http/dto/api.gen.go

.PHONY: generate lint lint-fix

generate: ## Цель для генерации DTO из openapi
	@oapi-codegen \
		-generate types \
		-package dto \
		-o $(GEN_OUTPUT) \
		$(OPENAPI_SPEC)

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

gen-dto: generate