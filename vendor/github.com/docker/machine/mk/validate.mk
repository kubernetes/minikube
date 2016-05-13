# Validate DCO on all history
mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

# XXX vendorized script miss exec bit, hence the gymnastic
# plus the path resolution...
# TODO migrate away from the shell script and have a make equivalent instead
dco:
	@echo `bash $(current_dir)/../script/validate-dco`

fmt:
	@test -z "$$(gofmt -s -l . 2>&1 | grep -v vendor/ | tee /dev/stderr)"

vet:
	@test -z "$$(go vet $(PKGS) 2>&1 | tee /dev/stderr)"

lint:
	$(if $(GOLINT), , \
		$(error Please install golint: go get -u github.com/golang/lint/golint))
	@test -z "$$($(GOLINT) ./... 2>&1 | grep -v vendor/ | grep -v "cli/" | grep -v "amazonec2/" |grep -v "openstack/" |grep -v "softlayer/" | grep -v "should have comment" | tee /dev/stderr)"
