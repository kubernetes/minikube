.PHONY: build-in-container build-local
DEPS:=$(wildcard sample/*.go) $(wildcard *.go) Dockerfile.build Makefile

build-in-container: $(DEPS) clean
	@echo "+ $@"
	@docker build -t hyperkitgo-build -f ./Dockerfile.build .
	@docker run --rm \
		-v ${CURDIR}:/go/src/github.com/moby/hyperkit/go \
		hyperkitgo-build build-local

build-local: build/hyperkitgo

build/hyperkitgo: $(DEPS)
	@echo "+ $@"
	GOOS=darwin GOARCH=amd64 \
	go build -o $@ --ldflags '-extldflags "-fno-PIC"' \
		sample/main.go

# The next two targets make sure we can cross-compile for Linux and Windows.
# While it makes little sense to run the sample code, it makes it easier for
# consumers of this package if it cross compiles.
build/hyperkitgo_linux: $(DEPS)
	@echo "+ $@"
	GOOS=linux GOARCH=amd64 \
	go build -o $@ --ldflags '-extldflags "-fno-PIC"' \
		sample/main.go
build/hyperkitgo.exe: $(DEPS)
	@echo "+ $@"
	GOOS=windows GOARCH=amd64 \
	go build -o $@ --ldflags '-extldflags "-fno-PIC"' \
		sample/main.go

clean:
	rm -rf build

fmt:
	@echo "+ $@"
	@gofmt -s -l . 2>&1 | grep -v ^vendor/ | xargs gofmt -s -l -w

lint:
	@echo "+ $@"
	$(if $(shell which golint || echo ''), , \
		$(error Please install golint))
	@test -z "$$(golint ./... 2>&1 | grep -v ^vendor/ | grep -v mock/ | tee /dev/stderr)"

# this will blow away the vendored code and update it to latest
.PHONY: vendor vendor-local
vendor:
	@echo "+ $@"
	@rm -rf vendor.conf vendor
	@docker build -t hyperkitgo-build -f ./Dockerfile.build .
	@docker run --rm \
		-v ${CURDIR}:/go/src/github.com/moby/hyperkit/go \
		hyperkitgo-build vendor-local VERSION=${VERSION} REVISION=${REVISION}

vendor-local:
	@echo "+ $@"
	@go get github.com/LK4D4/vndr
	@vndr init

.PHONY: ci setup
setup:
	go get github.com/golang/lint/golint

ci: setup build/hyperkitgo build/hyperkitgo_linux build/hyperkitgo.exe
	test -z "$$(gofmt -s -l . 2>&1 | grep -v ^vendor/ | tee /dev/stderr)"
	test -z "$$(golint ./... 2>&1 | grep -v ^vendor/ | grep -v mock/ | tee /dev/stderr)"
