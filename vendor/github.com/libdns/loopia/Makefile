
.PHONY: all
all: tidy build test audit
	@echo "all done"


.PHONY: build
build:
	@echo "building..."
	go build -v ./...


.PHONY: test
test: 
	go test ./...


.PHONY: tidy
tidy:
	@echo "tidy and fmt..."
	go mod tidy -v
	go fmt ./...


.PHONY: audit
audit:
	@echo "running audit checks..."
	go mod verify
	go vet ./...
	go list -m all
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...


no-dirty:
	@echo "checking for uncommitted changes..."
	git diff --exit-code
	git diff --cached --exit-code


.PHONY: example
example:
	go run ./_examples/up-and-down


.PHONY: acme-like
acme-like:
	go run ./_examples/acme-like


.PHONY: list-records
list-records:
	go run ./_examples/list-records
