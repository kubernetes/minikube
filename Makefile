# Copyright 2016 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Bump these on release
VERSION_MAJOR ?= 0
VERSION_MINOR ?= 20
VERSION_BUILD ?= 0
VERSION ?= v$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_BUILD)
DEB_VERSION ?= $(VERSION_MAJOR).$(VERSION_MINOR)-$(VERSION_BUILD)
INSTALL_SIZE ?= $(shell du out/minikube-windows-amd64.exe | cut -f1)
BUILDROOT_BRANCH ?= 2017.02
REGISTRY?=gcr.io/k8s-minikube

MINIKUBE_BUILD_IMAGE 	?= karalabe/xgo-1.8.3
LOCALKUBE_BUILD_IMAGE 	?= gcr.io/google_containers/kube-cross:v1.8.3-1
ISO_BUILD_IMAGE ?= $(REGISTRY)/buildroot-image

ISO_VERSION ?= v0.23.0
ISO_BUCKET ?= minikube/iso

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR ?= ./out
ORG := k8s.io
REPOPATH ?= $(ORG)/minikube

TAR_TARGETS_LINUX   := out/minikube-linux-amd64
TAR_TARGETS_DARWIN  := out/minikube-darwin-amd64
TAR_TARGETS_WINDOWS := out/minikube-windows-amd64.exe
TAR_TARGET_ALL      :=  $(TAR_TARGETS_LINUX) $(TAR_TARGETS_DARWIN) $(TAR_TARGETS_WINDOWS) 

# Use system python if it exists, otherwise use Docker.
PYTHON := $(shell command -v python || echo "docker run --rm -it -v $(shell pwd):/minikube -w /minikube python python")
BUILD_OS := $(shell uname -s)

LOCALKUBE_VERSION := $(shell $(PYTHON) hack/get_k8s_version.py --k8s-version-only 2>&1)
LOCALKUBE_BUCKET ?= minikube/k8sReleases
LOCALKUBE_UPLOAD_LOCATION := gs://${LOCALKUBE_BUCKET}
TAG ?= $(LOCALKUBE_VERSION)

# Set the version information for the Kubernetes servers, and build localkube statically
K8S_VERSION_LDFLAGS := $(shell $(PYTHON) hack/get_k8s_version.py 2>&1)
MINIKUBE_LDFLAGS := -X k8s.io/minikube/pkg/version.version=$(VERSION) -X k8s.io/minikube/pkg/version.isoVersion=$(ISO_VERSION) -X k8s.io/minikube/pkg/version.isoPath=$(ISO_BUCKET)
LOCALKUBE_LDFLAGS := "$(K8S_VERSION_LDFLAGS) $(MINIKUBE_LDFLAGS) -s -w -extldflags '-static'"

LOCALKUBEFILES := GOPATH=$(GOPATH) go list  -f '{{join .Deps "\n"}}' ./cmd/localkube/ | grep k8s.io | GOPATH=$(GOPATH) xargs go list -f '{{ range $$file := .GoFiles }} {{$$.Dir}}/{{$$file}}{{"\n"}}{{end}}'
MINIKUBEFILES := GOPATH=$(GOPATH) go list  -f '{{join .Deps "\n"}}' ./cmd/minikube/ | grep k8s.io | GOPATH=$(GOPATH) xargs go list -f '{{ range $$file := .GoFiles }} {{$$.Dir}}/{{$$file}}{{"\n"}}{{end}}'

MINIKUBE_ENV_LINUX   := CGO_ENABLED=1 GOARCH=amd64 GOOS=linux
MINIKUBE_ENV_DARWIN  := CGO_ENABLED=1 GOARCH=amd64 GOOS=darwin
MINIKUBE_ENV_WINDOWS := CGO_ENABLED=0 GOARCH=amd64 GOOS=windows

# extra env vars that need to be set in cross build container
MINIKUBE_ENV_DARWIN_DOCKER := CC=o64-clang CXX=o64-clang++

MINIKUBE_DOCKER_CMD := docker run -e IN_DOCKER=1 --user $(shell id -u):$(shell id -g) --workdir /go/src/$(REPOPATH) --entrypoint /bin/bash -v $(PWD):/go/src/$(REPOPATH) $(MINIKUBE_BUILD_IMAGE) -c
KUBE_CROSS_DOCKER_CMD := docker run -w /go/src/$(REPOPATH) --user $(shell id -u):$(shell id -g) -e IN_DOCKER=1 -v $(shell pwd):/go/src/$(REPOPATH) $(LOCALKUBE_BUILD_IMAGE)

# $(call MINIKUBE_GO_BUILD_CMD, output file, OS)
define MINIKUBE_GO_BUILD_CMD
	$($(shell echo MINIKUBE_ENV_$(2) | tr a-z A-Z)) go build --installsuffix cgo -ldflags="$(MINIKUBE_LDFLAGS) $(K8S_VERSION_LDFLAGS)" -a -o $(1) k8s.io/minikube/cmd/minikube
endef

ifeq ($(BUILD_IN_DOCKER),y)
	MINIKUBE_BUILD_IN_DOCKER=y
	LOCALKUBE_BUILD_IN_DOCKER=y
endif

# If not on linux, localkube must be built in docker
ifneq ($(BUILD_OS),Linux)
	LOCALKUBE_BUILD_IN_DOCKER=y
endif

# If we are already running in docker, 
# prevent recursion by unsetting the BUILD_IN_DOCKER directives.
# The _BUILD_IN_DOCKER variables should not be modified after this conditional.
ifeq ($(IN_DOCKER),1)
	MINIKUBE_BUILD_IN_DOCKER=n
	LOCALKUBE_BUILD_IN_DOCKER=n
endif

ifeq ($(GOOS),windows)
	IS_EXE = ".exe"
endif
out/minikube$(IS_EXE): gopath out/minikube-$(GOOS)-$(GOARCH)$(IS_EXE)
	cp $(BUILD_DIR)/minikube-$(GOOS)-$(GOARCH) $(BUILD_DIR)/minikube$(IS_EXE)

out/localkube: $(shell $(LOCALKUBEFILES))
ifeq ($(LOCALKUBE_BUILD_IN_DOCKER),y)
	$(KUBE_CROSS_DOCKER_CMD) make $@
else
	CGO_ENABLED=1 go build -tags static_build -ldflags=$(LOCALKUBE_LDFLAGS) -o $(BUILD_DIR)/localkube ./cmd/localkube
endif

out/minikube-windows-amd64.exe: out/minikube-windows-amd64
	mv out/minikube-windows-amd64 out/minikube-windows-amd64.exe

out/minikube-%-amd64: pkg/minikube/assets/assets.go $(shell $(MINIKUBEFILES))
ifeq ($(MINIKUBE_BUILD_IN_DOCKER),y)
	$(MINIKUBE_DOCKER_CMD) '$($(shell echo MINIKUBE_ENV_$*_DOCKER | tr a-z A-Z)) $(call MINIKUBE_GO_BUILD_CMD,$@,$*)'
else
	$(call MINIKUBE_GO_BUILD_CMD,$@,$*)
endif

minikube_iso: # old target kept for making tests happy
	echo $(ISO_VERSION) > deploy/iso/minikube-iso/board/coreos/minikube/rootfs-overlay/etc/VERSION
	if [ ! -d $(BUILD_DIR)/buildroot ]; then \
		mkdir -p $(BUILD_DIR); \
		git clone --branch=$(BUILDROOT_BRANCH) https://github.com/buildroot/buildroot $(BUILD_DIR)/buildroot; \
	fi;
	$(MAKE) BR2_EXTERNAL=../../deploy/iso/minikube-iso minikube_defconfig -C $(BUILD_DIR)/buildroot
	$(MAKE) -C $(BUILD_DIR)/buildroot
	mv $(BUILD_DIR)/buildroot/output/images/rootfs.iso9660 $(BUILD_DIR)/minikube.iso

out/minikube.iso: $(shell find deploy/iso/minikube-iso -type f)
ifeq ($(IN_DOCKER),1)
	$(MAKE) minikube_iso
else
	docker run --rm --workdir /mnt --volume $(CURDIR):/mnt $(ISO_DOCKER_EXTRA_ARGS) \
		--user $(shell id -u):$(shell id -g) --env HOME=/tmp --env IN_DOCKER=1 \
		$(ISO_BUILD_IMAGE) /usr/bin/make out/minikube.iso
endif

test-iso:
	go test -v $(REPOPATH)/test/integration --tags=iso --minikube-args="--iso-url=file://$(shell pwd)/out/buildroot/output/images/rootfs.iso9660"

.PHONY: integration
integration: out/minikube
	go test -v -test.timeout=30m $(REPOPATH)/test/integration --tags=integration --minikube-args="$(MINIKUBE_ARGS)"

.PHONY: integration-versioned
integration-versioned: out/minikube
	go test -v -test.timeout=30m $(REPOPATH)/test/integration --tags="integration versioned" --minikube-args="$(MINIKUBE_ARGS)"

.PHONY: test
test: pkg/minikube/assets/assets.go
	./test.sh

.PHONY: gopath
gopath:
ifneq ($(GOPATH)/src/$(REPOPATH),$(PWD))
	$(warning Warning: Building minikube outside the GOPATH, should be $(GOPATH)/src/$(REPOPATH) but is $(PWD))
endif

pkg/minikube/assets/assets.go: out/localkube $(GOPATH)/bin/go-bindata $(shell find deploy/addons -type f)
	$(GOPATH)/bin/go-bindata -nomemcopy -o pkg/minikube/assets/assets.go -pkg assets ./out/localkube deploy/addons/...

$(GOPATH)/bin/go-bindata:
	GOBIN=$(GOPATH)/bin go get github.com/jteeuwen/go-bindata/...

.PHONY: cross
cross: out/localkube out/minikube-linux-amd64 out/minikube-darwin-amd64 out/minikube-windows-amd64.exe

.PHONY: checksum
checksum:
	for f in out/localkube out/minikube-linux-amd64 out/minikube-darwin-amd64 out/minikube-windows-amd64.exe out/minikube.iso; do \
		if [ -f "$${f}" ]; then \
			openssl sha256 "$${f}" | awk '{print $$2}' > "$${f}.sha256" ; \
		fi ; \
	done

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -f pkg/minikube/assets/assets.go

.PHONY: gendocs
gendocs: out/docs/minikube.md

out/docs/minikube.md: $(shell find cmd) $(shell find pkg/minikube/constants) pkg/minikube/assets/assets.go
	cd $(GOPATH)/src/$(REPOPATH) && go run -ldflags="$(K8S_VERSION_LDFLAGS) $(MINIKUBE_LDFLAGS)" -tags gendocs hack/gen_help_text.go

out/minikube_$(DEB_VERSION).deb: out/minikube-linux-amd64
	cp -r installers/linux/deb/minikube_deb_template out/minikube_$(DEB_VERSION)
	chmod 0755 out/minikube_$(DEB_VERSION)/DEBIAN
	sed -E -i 's/--VERSION--/'$(DEB_VERSION)'/g' out/minikube_$(DEB_VERSION)/DEBIAN/control
	mkdir -p out/minikube_$(DEB_VERSION)/usr/bin
	cp out/minikube-linux-amd64 out/minikube_$(DEB_VERSION)/usr/bin/minikube
	dpkg-deb --build out/minikube_$(DEB_VERSION)
	rm -rf out/minikube_$(DEB_VERSION)


out/minikube-%-amd64.tar.gz:
	$(MAKE) $($(shell echo TAR_TARGETS_$* | tr a-z A-Z))
	tar -cvf $@ $($(shell echo TAR_TARGETS_$* | tr a-z A-Z))
	
out/minikube-installer.exe: out/minikube-windows-amd64.exe
	rm -rf out/windows_tmp
	cp -r installers/windows/ out/windows_tmp
	cp -r LICENSE out/windows_tmp/LICENSE
	awk 'sub("$$", "\r")' out/windows_tmp/LICENSE > out/windows_tmp/LICENSE.txt
	sed -E -i 's/--VERSION_MAJOR--/'$(VERSION_MAJOR)'/g' out/windows_tmp/minikube.nsi
	sed -E -i 's/--VERSION_MINOR--/'$(VERSION_MINOR)'/g' out/windows_tmp/minikube.nsi
	sed -E -i 's/--VERSION_BUILD--/'$(VERSION_BUILD)'/g' out/windows_tmp/minikube.nsi
	sed -E -i 's/--INSTALL_SIZE--/'$(INSTALL_SIZE)'/g' out/windows_tmp/minikube.nsi
	cp out/minikube-windows-amd64.exe out/windows_tmp/minikube.exe
	makensis out/windows_tmp/minikube.nsi
	mv out/windows_tmp/minikube-installer.exe out/minikube-installer.exe
	rm -rf out/windows_tmp

.PHONY: check-release
check-release:
	go test -v ./deploy/minikube/release_sanity_test.go -tags=release

.PHONY: release-localkube
release-localkube: out/localkube checksum
	gsutil cp out/localkube $(LOCALKUBE_UPLOAD_LOCATION)/$(LOCALKUBE_VERSION)/localkube-linux-amd64
	gsutil cp out/localkube.sha256 $(LOCALKUBE_UPLOAD_LOCATION)/$(LOCALKUBE_VERSION)/localkube-linux-amd64.sha256

.PHONY: update-releases
update-releases:
	gsutil cp deploy/minikube/k8s_releases.json gs://minikube/k8s_releases.json

localkube-image: out/localkube
	# TODO(aprindle) make addons placed into container configurable
	docker build -t $(REGISTRY)/localkube-image:$(TAG) -f deploy/docker/Dockerfile .
	@echo ""
	@echo "${REGISTRY}/localkube-image:$(TAG) succesfully built"
	@echo "See https://github.com/kubernetes/minikube/tree/master/deploy/docker for instructions on how to run image"

buildroot-image: $(ISO_BUILD_IMAGE) # convenient alias to build the docker container
$(ISO_BUILD_IMAGE): deploy/iso/minikube-iso/Dockerfile
	docker build $(ISO_DOCKER_EXTRA_ARGS) -t $@ -f $< $(dir $<)
	@echo ""
	@echo "$(@) successfully built"

.PHONY: release-iso
release-iso: minikube_iso checksum
	gsutil cp out/minikube.iso gs://$(ISO_BUCKET)/minikube-$(ISO_VERSION).iso
	gsutil cp out/minikube.iso.sha256 gs://$(ISO_BUCKET)/minikube-$(ISO_VERSION).iso.sha256
