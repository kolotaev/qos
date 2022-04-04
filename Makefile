OUT=./out
BINARY=server

default: run

deps:
	@go mod vendor

build:
	@go build -o $(OUT)/$(BINARY) ./example/main.go

run: build
	@$(OUT)/$(BINARY)

fmt:
	@go fmt ./...

test-unit:
	@go test . -v -run Test[^End2End]

test-e2e:
	@go test . -v -run End2End -timeout 120s

clean:
	@rm -rf $(OUT)
