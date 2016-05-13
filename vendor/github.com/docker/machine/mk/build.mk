extension = $(patsubst windows,.exe,$(filter windows,$(1)))

# Valid target combinations
VALID_OS_ARCH := "[darwin/amd64][linux/amd64][windows/amd64][windows/386]"

os.darwin := Darwin
os.linux := Linux
os.windows := Windows

arch.amd64 := x86_64
arch.386 := i386

define gocross
	$(if $(findstring [$(1)/$(2)],$(VALID_OS_ARCH)), \
	GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 \
		$(GO) build \
		-o $(PREFIX)/bin/docker-machine-${os.$(1)}-${arch.$(2)}$(call extension,$(GOOS)) \
		-a $(VERBOSE_GO) -tags "static_build netgo $(BUILDTAGS)" -installsuffix netgo \
		-ldflags "$(GO_LDFLAGS) -extldflags -static" $(GO_GCFLAGS) ./cmd/machine.go;)
endef

build-clean:
	rm -Rf $(PREFIX)/bin/*

build-x: $(shell find . -type f -name '*.go')
	$(foreach GOARCH,$(TARGET_ARCH),$(foreach GOOS,$(TARGET_OS),$(call gocross,$(GOOS),$(GOARCH))))

$(PREFIX)/bin/docker-machine$(call extension,$(GOOS)): $(shell find . -type f -name '*.go')
	$(GO) build \
	-o $@ \
	$(VERBOSE_GO) -tags "$(BUILDTAGS)" \
	-ldflags "$(GO_LDFLAGS)" $(GO_GCFLAGS) ./cmd/machine.go

build: $(PREFIX)/bin/docker-machine$(call extension,$(GOOS))
