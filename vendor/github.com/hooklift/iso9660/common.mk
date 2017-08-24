PLATFORM 	:= $(shell go env | grep GOHOSTOS | cut -d '"' -f 2)
ARCH 		:= $(shell go env | grep GOARCH | cut -d '"' -f 2)
BRANCH		:= $(shell git rev-parse --abbrev-ref HEAD)
LDFLAGS 	:= -ldflags "-X main.Version=$(VERSION) -X main.Name=$(NAME)"

test:
	go test ./...

build:
	go build -o build/$(NAME) $(LDFLAGS) cmd/$(NAME)/main.go

install:
	go install $(LDFLAGS)

compile:
	@rm -rf build/
	@gox $(LDFLAGS) \
	-osarch="darwin/amd64" \
	-os="linux" \
	-os="solaris" \
	-os="freebsd" \
	-os="windows" \
	-output "build/$(NAME)_$(VERSION)_{{.OS}}_{{.Arch}}/$(NAME)" \
	./...

dist: compile
	$(eval FILES := $(shell ls build))
	@rm -rf dist && mkdir dist
	@for f in $(FILES); do \
		(cd $(shell pwd)/build/$$f && tar -cvzf ../../dist/$$f.tar.gz *); \
		(cd $(shell pwd)/dist && shasum -a 512 $$f.tar.gz > $$f.sha512); \
		echo $$f; \
	done

release: dist
	@latest_tag=$$(git describe --tags `git rev-list --tags --max-count=1`); \
	comparison="$$latest_tag..HEAD"; \
	if [ -z "$$latest_tag" ]; then comparison=""; fi; \
	changelog=$$(git log $$comparison --oneline --no-merges); \
	github-release $(GHACCOUNT)/$(NAME) $(VERSION) $(BRANCH) "**Changelog**<br/>$$changelog" 'dist/*'; \
	git pull

.PHONY: test build install compile deps dist release
