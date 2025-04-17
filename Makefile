default: build
	./dnsclay serve -loglevel debug -trace text,json -dns-notify-tcpaddr localhost:1053 -dns-notify-tlsaddr localhost:1853

build: node_modules/.bin/tsc providers.go
	CGO_ENABLED=0 go build
	./gendoc.sh
	./genapidoc.sh
	./gents.sh web/api.json api.ts
	./tsc.sh web/dnsclay.js dom.ts api.ts dnsclay.ts
	CGO_ENABLED=0 go build # build with generated files

providers.go: providers.txt Makefile genapidoc.sh genproviders.sh
	./genproviders.sh
	go mod tidy
	go mod vendor
	./genlicense.sh

race:
	go build -race

check:
	GOARCH=386 CGO_ENABLED=0 go vet
	CGO_ENABLED=0 staticcheck

check-shadow:
	go vet -vettool=$$(which shadow) ./... 2>&1 | grep -v '"err"'

# for debugging or dns tracing:
# DNSCLAY_TEST_LOGLEVEL=debug
# DNSCLAY_TEST_LOGLEVEL=debug-1
test:
	CGO_ENABLED=0 go test -shuffle=on -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html

test-race:
	CGO_ENABLED=1 go test -race -shuffle=on

knotd:
	-mkdir testdata/knot/run testdata/knot/db testdata/knot/var
	test -f testdata/knot/var/example.com.zone || cp testdata/knot/example.com.zone.initial testdata/knot/var/example.com.zone
	/usr/sbin/knotd -c testdata/knot/knot.conf -v

tswatch:
	bash -c 'while true; do inotifywait -q -e close_write *.ts; make web/dnsclay.js; done'

frontend:
	./tsc.sh web/dnsclay.js dom.ts api.ts dnsclay.ts

web/dnsclay.js: node_modules/.bin/tsc dom.ts api.ts dnsclay.ts
	./tsc.sh web/dnsclay.js dom.ts api.ts dnsclay.ts

node_modules/.bin/tsc:
	-mkdir -p node_modules/.bin
	npm ci --ignore-scripts

install-js:
	-mkdir -p node_modules/.bin
	npm install --ignore-scripts --save-dev --save-exact typescript@5.1.6

govendor:
	go mod tidy
	go mod vendor
	./genlicense.sh

buildall:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build
	CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build
	CGO_ENABLED=0 GOOS=netbsd GOARCH=amd64 go build
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build
	CGO_ENABLED=0 GOOS=dragonfly GOARCH=amd64 go build
	CGO_ENABLED=0 GOOS=illumos GOARCH=amd64 go build
	CGO_ENABLED=0 GOOS=solaris GOARCH=amd64 go build
	CGO_ENABLED=0 GOOS=aix GOARCH=ppc64 go build
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
	# no bbolt: CGO_ENABLED=0 GOOS=plan9 GOARCH=amd64 go build

fmt:
	go fmt ./...
	gofmt -w -s *.go

clean:
	CGO_ENABLED=0 go clean
