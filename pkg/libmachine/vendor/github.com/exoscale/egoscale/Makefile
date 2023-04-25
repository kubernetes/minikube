PKG=github.com/exoscale/egoscale

GOPATH=$(CURDIR)/.gopath
DEP=$(GOPATH)/bin/dep

export GOPATH

$(GOPATH)/src/$(PKG):
	mkdir -p $(GOPATH)
	go get -u github.com/golang/dep/cmd/dep
	mkdir -p $(shell dirname $(GOPATH)/src/$(PKG))
	ln -sf ../../../.. $(GOPATH)/src/$(PKG)

.PHONY: deps
deps: $(GOPATH)/src/$(PKG)
	(cd $(GOPATH)/src/$(PKG) && \
		$(DEP) ensure)

.PHONY: deps-update
deps-update: deps
	(cd $(GOPATH)/src/$(PKG) && \
		$(DEP) ensure -update)

.PHONY: clean
clean:
	$(RM) -r $(DEST)
	go clean
