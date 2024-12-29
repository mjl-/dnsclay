
.PHONY: all tidy test audit

all: tidy test audit


tidy:
	go fmt ./...
	go mod tidy -v


test:
	@echo "Testing..."
	go test -v -buildvcs -count=1 ./...


audit:
	go mod verify
	go vet ./...

	go list -m all
	go list -m -u all
	
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...


