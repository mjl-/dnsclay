go_bin:=$(shell pwd)/bin

# if you want to implement api updates then use bash `make`
# and extend/update/fix sdk code after

all: clean tools dep lint

tools:
	GOBIN=${go_bin} go install -mod=mod github.com/golangci/golangci-lint/cmd/golangci-lint@v1.39.0

clean:
	rm -rf ./bin
	mkdir ./bin

dep:
	go mod tidy
	go mod vendor

updatedep:
	go list -m -u all

lint:
	./bin/golangci-lint run ./client*.go

test:
	go test --count=1 -v -race ./...

.PHONY: lint test updatedep