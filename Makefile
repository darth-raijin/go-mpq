.PHONY: generate
generate:
	@echo "Generating proto files"
	protoc -I=./internal/protos --go_out=./internal/protos ./internal/protos/*.proto
	@echo "Generated proto files"