IGNORE_FILE = $(foreach file,Makefile,--ignore $(file))
IGNORE_DIR = $(foreach dir,vendor testdata internal,--ignore-dir $(dir))
IGNORE = $(IGNORE_FILE) $(IGNORE_DIR)

all: test

install:
	go install -v -x ./...

install-force:
	rm -rf $(GOPATH)/pkg/darwin_amd64/github.com/zchee/go-qcow2 $(GOPATH)/pkg/darwin_amd64/github.com/zchee/go-qcow2.a
	go install -v -x ./...

test:
	go run cmd/qcow-test/qcow-test.go
	readbyte search testdata/test.qcow2

todo: 
	@ag 'TODO(\(.+\):|:)' --after=1 $(IGNORE) || true
	@ag 'BUG(\(.+\):|:)' --after=1 $(IGNORE)|| true
	@ag 'XXX(\(.+\):|:)' --after=1 $(IGNORE)|| true
	@ag 'FIXME(\(.+\):|:)' --after=1 $(IGNORE) || true
	@ag 'NOTE(\(.+\):|:)' --after=1 $(IGNORE) || true

.PHONY: todo test install
