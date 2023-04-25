# COVERAGE_OUTPUT dir is a temp dir (OSX/Linux compatible), unless explicitly specified through env COVERAGE_DIR
COVERAGE_OUTPUT := $(COVERAGE_DIR)
ifeq ($(COVERAGE_OUTPUT),)
	COVERAGE_OUTPUT := $(shell mktemp -d 2>/dev/null || mktemp -d -t machine-coverage)
endif

# Final cover file, html, and mode
COVERAGE_PROFILE := $(COVERAGE_OUTPUT)/profile.out
COVERAGE_HTML := $(COVERAGE_OUTPUT)/index.html
COVERAGE_MODE := set

# Goveralls dependency
GOVERALLS_BIN := $(GOPATH)/bin/goveralls
GOVERALLS := $(shell [ -x $(GOVERALLS_BIN) ] && echo $(GOVERALLS_BIN) || echo '')

# Generate coverage
coverage-generate: $(COVERAGE_PROFILE)

# Send the results to coveralls
coverage-send: $(COVERAGE_PROFILE)
	$(if $(GOVERALLS), , $(error Please install goveralls: go get github.com/mattn/goveralls))
	@$(GOVERALLS) -service travis-ci -coverprofile="$(COVERAGE_PROFILE)"

# Generate html report
coverage-html: $(COVERAGE_HTML)
	@open "$(COVERAGE_HTML)"

# Serve over http - useful only if building remote/headless
coverage-serve: $(COVERAGE_HTML)
	@cd "$(COVERAGE_OUTPUT)" && python -m SimpleHTTPServer 8000

# Clean up coverage coverage output
coverage-clean:
	@rm -Rf "$(COVERAGE_OUTPUT)/coverage"
	@rm -f "$(COVERAGE_HTML)"
	@rm -f "$(COVERAGE_PROFILE)"

$(COVERAGE_PROFILE): $(shell find . -type f -name '*.go')
	@mkdir -p "$(COVERAGE_OUTPUT)/coverage"
	@$(foreach PKG,$(PKGS), go test $(VERBOSE_GO) -tags "$(BUILDTAGS)" -covermode=$(COVERAGE_MODE) -coverprofile="$(COVERAGE_OUTPUT)/coverage/`echo $(PKG) | tr "/" "-"`.cover" "$(PKG)";)
	@echo "mode: $(COVERAGE_MODE)" > "$(COVERAGE_PROFILE)"
	@grep -h -v "^mode:" "$(COVERAGE_OUTPUT)/coverage"/*.cover >> "$(COVERAGE_PROFILE)"

$(COVERAGE_HTML): $(COVERAGE_PROFILE)
	$(GO) tool cover -html="$(COVERAGE_PROFILE)" -o "$(COVERAGE_HTML)"
