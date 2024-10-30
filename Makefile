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

# Bump these on release - and please check ISO_VERSION for correctness.
VERSION_MAJOR ?= 1
VERSION_MINOR ?= 34
VERSION_BUILD ?= 0
RAW_VERSION=$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_BUILD)
VERSION ?= v$(RAW_VERSION)

KUBERNETES_VERSION ?= $(shell grep -E "DefaultKubernetesVersion =" pkg/minikube/constants/constants.go | cut -d \" -f2)
KIC_VERSION ?= $(shell grep -E "Version =" pkg/drivers/kic/types.go | cut -d \" -f2)
HUGO_VERSION ?= $(shell grep -E "HUGO_VERSION = \"" netlify.toml | cut -d \" -f2)

# Default to .0 for higher cache hit rates, as build increments typically don't require new ISO versions
ISO_VERSION ?= v1.34.0-1730282777-19883

# Dashes are valid in semver, but not Linux packaging. Use ~ to delimit alpha/beta
DEB_VERSION ?= $(subst -,~,$(RAW_VERSION))
DEB_REVISION ?= 0

RPM_VERSION ?= $(DEB_VERSION)
RPM_REVISION ?= 0

# used by hack/jenkins/release_build_and_upload.sh and KVM_BUILD_IMAGE, see also BUILD_IMAGE below
# update this only by running `make update-golang-version`
GO_VERSION ?= 1.23.2
# update this only by running `make update-golang-version`
GO_K8S_VERSION_PREFIX ?= v1.32.0

# replace "x.y.0" => "x.y". kube-cross and go.dev/dl use different formats for x.y.0 go versions
KVM_GO_VERSION ?= $(GO_VERSION:.0=)


INSTALL_SIZE ?= $(shell du out/minikube-windows-amd64.exe | cut -f1)
BUILDROOT_BRANCH ?= 2023.02.9
# the go version on the line below is for the ISO
GOLANG_OPTIONS = GO_VERSION=1.21.6 GO_HASH_FILE=$(PWD)/deploy/iso/minikube-iso/go.hash
BUILDROOT_OPTIONS = BR2_EXTERNAL=../../deploy/iso/minikube-iso $(GOLANG_OPTIONS)
REGISTRY ?= gcr.io/k8s-minikube

# Get git commit id
COMMIT_NO := $(shell git rev-parse HEAD 2> /dev/null || true)
COMMIT ?= $(if $(shell git status --porcelain --untracked-files=no),"${COMMIT_NO}-dirty","${COMMIT_NO}")
COMMIT_SHORT = $(shell git rev-parse --short HEAD 2> /dev/null || true)
COMMIT_NOQUOTES := $(patsubst "%",%,$(COMMIT))
# source code for image: https://github.com/neilotoole/xcgo
HYPERKIT_BUILD_IMAGE ?= gcr.io/k8s-minikube/xcgo:go1.22.2

# NOTE: "latest" as of 2021-02-06. kube-cross images aren't updated as often as Kubernetes
# https://github.com/kubernetes/kubernetes/blob/master/build/build-image/cross/VERSION
#

BUILD_IMAGE 	?= registry.k8s.io/build-image/kube-cross:$(GO_K8S_VERSION_PREFIX)-go$(GO_VERSION)-bullseye.0

ISO_BUILD_IMAGE ?= $(REGISTRY)/buildroot-image

KVM_BUILD_IMAGE_AMD64 ?= $(REGISTRY)/kvm-build-image_amd64:$(KVM_GO_VERSION)
KVM_BUILD_IMAGE_ARM64 ?= $(REGISTRY)/kvm-build-image_arm64:$(KVM_GO_VERSION)

ISO_BUCKET ?= minikube/iso

MINIKUBE_VERSION ?= $(ISO_VERSION)
MINIKUBE_BUCKET ?= minikube/releases
MINIKUBE_UPLOAD_LOCATION := gs://${MINIKUBE_BUCKET}
MINIKUBE_RELEASES_URL=https://github.com/kubernetes/minikube/releases/download

KERNEL_VERSION ?= 5.10.207
# latest from https://github.com/golangci/golangci-lint/releases
# update this only by running `make update-golint-version`
GOLINT_VERSION ?= v1.61.0
# Limit number of default jobs, to avoid the CI builds running out of memory
GOLINT_JOBS ?= 4
# see https://github.com/golangci/golangci-lint#memory-usage-of-golangci-lint
GOLINT_GOGC ?= 100
# options for lint (golangci-lint)
GOLINT_OPTIONS = --timeout 7m \
	  --build-tags "${MINIKUBE_INTEGRATION_BUILD_TAGS}" \
	  --enable gofmt,goimports,gocritic,revive,gocyclo,misspell,nakedret,stylecheck,unconvert,unparam,dogsled \
	  --exclude 'variable on range scope.*in function literal|ifElseChain'

export GO111MODULE := on

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOARM ?= 7 # the default is 5
GOPATH ?= $(shell go env GOPATH)
BUILD_DIR ?= $(PWD)/out
$(shell mkdir -p $(BUILD_DIR))
CURRENT_GIT_BRANCH ?= $(shell git branch | grep \* | cut -d ' ' -f2)

# Use system python if it exists, otherwise use Docker.
PYTHON := $(shell command -v python || echo "docker run --rm -it -v $(shell pwd):/minikube -w /minikube python python")
BUILD_OS := $(shell uname -s)

SHA512SUM=$(shell command -v sha512sum || echo "shasum -a 512")

# check which "flavor" of SED is being used as the flags are different between BSD and GNU sed.
# BSD sed does not support "--version"
SED_VERSION := $(shell sed --version 2>/dev/null | head -n 1 | cut -d' ' -f4)
ifeq ($(SED_VERSION),)
	SED = sed -i ''
else
	SED = sed -i
endif

# gvisor tag to automatically push changes to
# to update minikubes default, update deploy/addons/gvisor
GVISOR_TAG ?= v0.0.2

# auto-pause-hook tag to push changes to
AUTOPAUSE_HOOK_TAG ?= v0.0.5

# prow-test tag to push changes to
PROW_TEST_TAG ?= v0.0.7

# storage provisioner tag to push changes to
# NOTE: you will need to bump the PreloadVersion if you change this
STORAGE_PROVISIONER_TAG ?= v5

STORAGE_PROVISIONER_MANIFEST ?= $(REGISTRY)/storage-provisioner:$(STORAGE_PROVISIONER_TAG)
STORAGE_PROVISIONER_IMAGE ?= $(REGISTRY)/storage-provisioner-$(GOARCH):$(STORAGE_PROVISIONER_TAG)

# Set the version information for the Kubernetes servers
MINIKUBE_LDFLAGS := -X k8s.io/minikube/pkg/version.version=$(VERSION) -X k8s.io/minikube/pkg/version.isoVersion=$(ISO_VERSION) -X k8s.io/minikube/pkg/version.gitCommitID=$(COMMIT) -X k8s.io/minikube/pkg/version.storageProvisionerVersion=$(STORAGE_PROVISIONER_TAG)
PROVISIONER_LDFLAGS := "-X k8s.io/minikube/pkg/storage.version=$(STORAGE_PROVISIONER_TAG) -s -w -extldflags '-static'"

MINIKUBEFILES := ./cmd/minikube/
HYPERKIT_FILES := ./cmd/drivers/hyperkit
STORAGE_PROVISIONER_FILES := ./cmd/storage-provisioner
KVM_DRIVER_FILES := ./cmd/drivers/kvm/

MINIKUBE_TEST_FILES := ./cmd/... ./pkg/...

# npm install -g markdownlint-cli
MARKDOWNLINT ?= markdownlint


MINIKUBE_MARKDOWN_FILES := README.md CONTRIBUTING.md CHANGELOG.md

MINIKUBE_BUILD_TAGS :=
MINIKUBE_INTEGRATION_BUILD_TAGS := integration $(MINIKUBE_BUILD_TAGS)

CMD_SOURCE_DIRS = cmd pkg deploy/addons translations
SOURCE_DIRS = $(CMD_SOURCE_DIRS) test
SOURCE_PACKAGES = ./cmd/... ./pkg/... ./deploy/addons/... ./translations/... ./test/...

SOURCE_FILES = $(shell find $(CMD_SOURCE_DIRS) -type f -name "*.go" | grep -v _test.go)
GOTEST_FILES = $(shell find $(CMD_SOURCE_DIRS) -type f -name "*.go" | grep _test.go)
ADDON_FILES = $(shell find "deploy/addons" -type f | grep -v "\.go")
TRANSLATION_FILES = $(shell find "translations" -type f | grep -v "\.go")
ASSET_FILES = $(ADDON_FILES) $(TRANSLATION_FILES)

# kvm2 ldflags
KVM2_LDFLAGS := -X k8s.io/minikube/pkg/drivers/kvm.version=$(VERSION) -X k8s.io/minikube/pkg/drivers/kvm.gitCommitID=$(COMMIT)

# hyperkit ldflags
HYPERKIT_LDFLAGS := -X k8s.io/minikube/pkg/drivers/hyperkit.version=$(VERSION) -X k8s.io/minikube/pkg/drivers/hyperkit.gitCommitID=$(COMMIT)

# autopush artefacts
AUTOPUSH ?=

# version file json
VERSION_JSON := "{\"iso_version\": \"$(ISO_VERSION)\", \"kicbase_version\": \"$(KIC_VERSION)\", \"minikube_version\": \"$(VERSION)\", \"commit\": \"$(COMMIT_NOQUOTES)\"}"

# don't ask for user confirmation
IN_CI := false

# $(call user_confirm, message)
define user_confirm
	@if [ "${IN_CI}" = "false" ]; then\
		echo "⚠️ $(1)";\
		read -p "Do you want to proceed? (Y/N): " confirm && echo $$confirm | grep -iq "^[yY]" || exit 1;\
	fi
endef

# $(call DOCKER, image, command)
define DOCKER
	docker run --rm -e GOCACHE=/app/.cache -e IN_DOCKER=1 --user $(shell id -u):$(shell id -g) -w /app -v $(PWD):/app -v $(GOPATH):/go --init $(1) /bin/bash -c '$(2)'
endef

ifeq ($(BUILD_IN_DOCKER),y)
	MINIKUBE_BUILD_IN_DOCKER=y
endif

# If we are already running in docker,
# prevent recursion by unsetting the BUILD_IN_DOCKER directives.
# The _BUILD_IN_DOCKER variables should not be modified after this conditional.
ifeq ($(IN_DOCKER),1)
	MINIKUBE_BUILD_IN_DOCKER=n
endif

ifeq ($(GOOS),windows)
	IS_EXE = .exe
	DIRSEP_ = \\
	DIRSEP = $(strip $(DIRSEP_))
	PATHSEP = ;
else
	DIRSEP = /
	PATHSEP = :
endif

v_at_0 = yes
v_at_ = $(v_at_1)
quiet := $(v_at_$(V))
Q=$(if $(quiet),@)

INTEGRATION_TESTS_TO_RUN := ./test/integration
ifneq ($(TEST_FILES),)
	TEST_HELPERS = main_test.go util_test.go helpers_test.go
	INTEGRATION_TESTS_TO_RUN := $(addprefix ./test/integration/, $(TEST_HELPERS) $(TEST_FILES))
endif

out/minikube$(IS_EXE): $(SOURCE_FILES) $(ASSET_FILES) go.mod
ifeq ($(MINIKUBE_BUILD_IN_DOCKER),y)
	$(call DOCKER,$(BUILD_IMAGE),GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) /usr/bin/make $@)
else
	$(if $(quiet),@echo "  GO       $@")
	$(Q)go build $(MINIKUBE_GOFLAGS) -tags "$(MINIKUBE_BUILD_TAGS)" -ldflags="$(MINIKUBE_LDFLAGS)" -o $@ k8s.io/minikube/cmd/minikube
endif

out/minikube-windows-amd64.exe: out/minikube-windows-amd64
	$(if $(quiet),@echo "  CP       $@")
	$(Q)cp $< $@

out/minikube-linux-i686: out/minikube-linux-386
	$(if $(quiet),@echo "  CP       $@")
	$(Q)cp $< $@

out/minikube-linux-x86_64: out/minikube-linux-amd64
	$(if $(quiet),@echo "  CP       $@")
	$(Q)cp $< $@

out/minikube-linux-armhf: out/minikube-linux-arm
	$(if $(quiet),@echo "  CP       $@")
	$(Q)cp $< $@

out/minikube-linux-armv7hl: out/minikube-linux-arm
	$(if $(quiet),@echo "  CP       $@")
	$(Q)cp $< $@

out/minikube-linux-aarch64: out/minikube-linux-arm64
	$(if $(quiet),@echo "  CP       $@")
	$(Q)cp $< $@

out/minikube-linux-ppc64el: out/minikube-linux-ppc64le
	$(if $(quiet),@echo "  CP       $@")
	$(Q)cp $< $@

.PHONY: minikube-linux-amd64 minikube-linux-arm64
minikube-linux-amd64: out/minikube-linux-amd64 ## Build Minikube for Linux x86 64bit
minikube-linux-arm64: out/minikube-linux-arm64 ## Build Minikube for Linux ARM 64bit

.PHONY: minikube-darwin-amd64 minikube-darwin-arm64
minikube-darwin-amd64: out/minikube-darwin-amd64 ## Build Minikube for Darwin x86 64bit
minikube-darwin-arm64: out/minikube-darwin-arm64 ## Build Minikube for Darwin ARM 64bit

.PHONY: minikube-windows-amd64.exe
minikube-windows-amd64.exe: out/minikube-windows-amd64.exe ## Build Minikube for Windows 64bit

eq = $(and $(findstring x$(1),x$(2)),$(findstring x$(2),x$(1)))

out/minikube-%: $(SOURCE_FILES) $(ASSET_FILES)
ifeq ($(MINIKUBE_BUILD_IN_DOCKER),y)
	$(call DOCKER,$(BUILD_IMAGE),/usr/bin/make $@)
else
	$(if $(quiet),@echo "  GO       $@")
	$(Q)GOOS="$(firstword $(subst -, ,$*))" GOARCH="$(lastword $(subst -, ,$(subst $(IS_EXE), ,$*)))" $(if $(call eq,$(lastword $(subst -, ,$(subst $(IS_EXE), ,$*))),arm),GOARM=$(GOARM)) \
	go build -tags "$(MINIKUBE_BUILD_TAGS)" -ldflags="$(MINIKUBE_LDFLAGS)" -a -o $@ k8s.io/minikube/cmd/minikube
endif

out/minikube-linux-armv6: $(SOURCE_FILES) $(ASSET_FILES)
	$(Q)GOOS=linux GOARCH=arm GOARM=6 \
	go build -tags "$(MINIKUBE_BUILD_TAGS)" -ldflags="$(MINIKUBE_LDFLAGS)" -a -o $@ k8s.io/minikube/cmd/minikube

.PHONY: e2e-linux-amd64 e2e-linux-arm64 e2e-darwin-amd64 e2e-darwin-arm64 e2e-windows-amd64.exe
e2e-linux-amd64: out/e2e-linux-amd64 ## build end2end binary for Linux x86 64bit
e2e-linux-arm64: out/e2e-linux-arm64 ## build end2end binary for Linux ARM 64bit
e2e-darwin-amd64: out/e2e-darwin-amd64 ## build end2end binary for Darwin x86 64bit
e2e-darwin-arm64: out/e2e-darwin-arm64 ## build end2end binary for Darwin ARM 64bit
e2e-windows-amd64.exe: out/e2e-windows-amd64.exe ## build end2end binary for Windows 64bit

out/e2e-%: out/minikube-%
	GOOS="$(firstword $(subst -, ,$*))" GOARCH="$(lastword $(subst -, ,$(subst $(IS_EXE), ,$*)))" go test -ldflags="${MINIKUBE_LDFLAGS}" -c k8s.io/minikube/test/integration --tags="$(MINIKUBE_INTEGRATION_BUILD_TAGS)" -o $@

out/e2e-windows-amd64.exe: out/e2e-windows-amd64
	cp $< $@

minikube-iso-amd64: minikube-iso-x86_64
minikube-iso-arm64: minikube-iso-aarch64

minikube-iso-%: deploy/iso/minikube-iso/board/minikube/%/rootfs-overlay/usr/bin/auto-pause # build minikube iso
	echo $(VERSION_JSON) > deploy/iso/minikube-iso/board/minikube/$*/rootfs-overlay/version.json
	echo $(ISO_VERSION) > deploy/iso/minikube-iso/board/minikube/$*/rootfs-overlay/etc/VERSION
	cp deploy/iso/minikube-iso/arch/$*/Config.in.tmpl deploy/iso/minikube-iso/Config.in
	if [ ! -d $(BUILD_DIR)/buildroot ]; then \
		mkdir -p $(BUILD_DIR); \
		git clone --depth=1 --branch=$(BUILDROOT_BRANCH) https://github.com/buildroot/buildroot $(BUILD_DIR)/buildroot; \
		perl -pi -e 's@\s+source "package/sysdig/Config\.in"\n@@;' $(BUILD_DIR)/buildroot/package/Config.in; \
		rm -r $(BUILD_DIR)/buildroot/package/sysdig; \
		cp deploy/iso/minikube-iso/go.hash $(BUILD_DIR)/buildroot/package/go/go.hash; \
		git --git-dir=$(BUILD_DIR)/buildroot/.git config user.email "dev@random.com"; \
		git --git-dir=$(BUILD_DIR)/buildroot/.git config user.name "Random developer"; \
	fi;
	$(MAKE) -C $(BUILD_DIR)/buildroot $(BUILDROOT_OPTIONS) O=$(BUILD_DIR)/buildroot/output-$* minikube_$*_defconfig
	$(MAKE) -C $(BUILD_DIR)/buildroot $(BUILDROOT_OPTIONS) O=$(BUILD_DIR)/buildroot/output-$* host-python3
	$(MAKE) -C $(BUILD_DIR)/buildroot $(BUILDROOT_OPTIONS) O=$(BUILD_DIR)/buildroot/output-$*
	# x86_64 ISO is still BIOS rather than EFI because of AppArmor issues for KVM, and Gen 2 issues for Hyper-V
	if [ "$*" = "aarch64" ]; then \
                mv $(BUILD_DIR)/buildroot/output-aarch64/images/boot.iso $(BUILD_DIR)/minikube-arm64.iso; \
        else \
                mv $(BUILD_DIR)/buildroot/output-x86_64/images/rootfs.iso9660 $(BUILD_DIR)/minikube-amd64.iso; \
        fi;

# Change buildroot configuration for the minikube ISO
.PHONY: iso-menuconfig
iso-menuconfig-%: ## Configure buildroot configuration
	$(MAKE) -C $(BUILD_DIR)/buildroot $(BUILDROOT_OPTIONS) O=$(BUILD_DIR)/buildroot/output-$* menuconfig
	$(MAKE) -C $(BUILD_DIR)/buildroot $(BUILDROOT_OPTIONS) O=$(BUILD_DIR)/buildroot/output-$* savedefconfig

# Change the kernel configuration for the minikube ISO
linux-menuconfig-%:  ## Configure Linux kernel configuration
	$(MAKE) -C $(BUILD_DIR)/buildroot/output-$*/build/linux-$(KERNEL_VERSION)/ menuconfig
	$(MAKE) -C $(BUILD_DIR)/buildroot/output-$*/build/linux-$(KERNEL_VERSION)/ savedefconfig
	cp $(BUILD_DIR)/buildroot/output-$*/build/linux-$(KERNEL_VERSION)/defconfig deploy/iso/minikube-iso/board/minikube/$*/linux_$*_defconfig

out/minikube-%.iso: $(shell find "deploy/iso/minikube-iso" -type f)
ifeq ($(IN_DOCKER),1)
	$(MAKE) minikube-iso-$*
else
	docker run --rm --workdir /mnt --volume $(CURDIR):/mnt $(ISO_DOCKER_EXTRA_ARGS) \
		--user $(shell id -u):$(shell id -g) --env HOME=/tmp --env IN_DOCKER=1 \
		$(ISO_BUILD_IMAGE) /bin/bash -lc '/usr/bin/make minikube-iso-$*'
endif

iso_in_docker:
	docker run -it --rm --workdir /mnt --volume $(CURDIR):/mnt $(ISO_DOCKER_EXTRA_ARGS) \
		--user $(shell id -u):$(shell id -g) --env HOME=/tmp --env IN_DOCKER=1 \
		$(ISO_BUILD_IMAGE) /bin/bash

.PHONY: test-pkg
test-pkg/%: ## Trigger packaging test
	go test -v -test.timeout=60m ./$* --tags="$(MINIKUBE_BUILD_TAGS)"

.PHONY: all
all: cross drivers e2e-cross cross-tars exotic retro out/gvisor-addon ## Build all different minikube components

.PHONY: drivers
drivers: ## Build Hyperkit and KVM2 drivers
drivers: docker-machine-driver-hyperkit \
	 docker-machine-driver-kvm2 \
	 out/docker-machine-driver-kvm2-amd64 \
	 out/docker-machine-driver-kvm2-arm64


.PHONY: docker-machine-driver-hyperkit
docker-machine-driver-hyperkit: out/docker-machine-driver-hyperkit ## Build Hyperkit driver

.PHONY: docker-machine-driver-kvm2
docker-machine-driver-kvm2: out/docker-machine-driver-kvm2 ## Build KVM2 driver

.PHONY: integration
integration: out/minikube$(IS_EXE) ## Trigger minikube integration test, logs to ./out/testout_COMMIT.txt
	go test -ldflags="${MINIKUBE_LDFLAGS}" -v -test.timeout=90m $(INTEGRATION_TESTS_TO_RUN) --tags="$(MINIKUBE_INTEGRATION_BUILD_TAGS)" $(TEST_ARGS) 2>&1 | tee "./out/testout_$(COMMIT_SHORT).txt"

.PHONY: integration-none-driver
integration-none-driver: e2e-linux-$(GOARCH) out/minikube-linux-$(GOARCH)  ## Trigger minikube none driver test, logs to ./out/testout_COMMIT.txt
	out/e2e-linux-$(GOARCH) -testdata-dir "test/integration/testdata" -minikube-start-args="--driver=none" -test.v -test.timeout=60m -binary=out/minikube-linux-amd64 $(TEST_ARGS) 2>&1 | tee "./out/testout_$(COMMIT_SHORT).txt"

.PHONY: integration-versioned
integration-versioned: out/minikube ## Trigger minikube integration testing, logs to ./out/testout_COMMIT.txt
	go test -ldflags="${MINIKUBE_LDFLAGS}" -v -test.timeout=90m $(INTEGRATION_TESTS_TO_RUN) --tags="$(MINIKUBE_INTEGRATION_BUILD_TAGS) versioned" $(TEST_ARGS) 2>&1 | tee "./out/testout_$(COMMIT_SHORT).txt"

.PHONY: functional
functional: integration-functional-only

.PHONY: integration-functional-only
integration-functional-only: out/minikube$(IS_EXE) ## Trigger only functioanl tests in integration test, logs to ./out/testout_COMMIT.txt
	go test -ldflags="${MINIKUBE_LDFLAGS}" -v -test.timeout=20m $(INTEGRATION_TESTS_TO_RUN) --tags="$(MINIKUBE_INTEGRATION_BUILD_TAGS)" $(TEST_ARGS) -test.run TestFunctional 2>&1 | tee "./out/testout_$(COMMIT_SHORT).txt"

.PHONY: html_report
html_report: ## Generate HTML  report out of the last ran integration test logs.
	@go tool test2json -t < "./out/testout_$(COMMIT_SHORT).txt" > "./out/testout_$(COMMIT_SHORT).json"
	@gopogh -in "./out/testout_$(COMMIT_SHORT).json" -out ./out/testout_$(COMMIT_SHORT).html -name "$(shell git rev-parse --abbrev-ref HEAD)" -pr "" -repo github.com/kubernetes/minikube/  -details "${COMMIT_SHORT}"
	@echo "-------------------------- Open HTML Report in Browser: ---------------------------"
ifeq ($(GOOS),windows)
	@echo start $(CURDIR)/out/testout_$(COMMIT_SHORT).html
	@echo "-----------------------------------------------------------------------------------"
	@start $(CURDIR)/out/testout_$(COMMIT_SHORT).html || true
else
	@echo open $(CURDIR)/out/testout_$(COMMIT_SHORT).html
	@echo "-----------------------------------------------------------------------------------"
	@open $(CURDIR)/out/testout_$(COMMIT_SHORT).html || true
endif

.PHONY: test
test: ## Trigger minikube test
	MINIKUBE_LDFLAGS="${MINIKUBE_LDFLAGS}" ./test.sh

.PHONY: generate-docs
generate-docs: extract out/minikube ## Automatically generate commands documentation.
	out/minikube generate-docs --path ./site/content/en/docs/commands/ --test-path ./site/content/en/docs/contrib/tests.en.md --code-path ./site/content/en/docs/contrib/errorcodes.en.md

.PHONY: gotest
gotest: ## Trigger minikube test
	$(if $(quiet),@echo "  TEST     $@")
	$(Q)go test -tags "$(MINIKUBE_BUILD_TAGS)" -ldflags="$(MINIKUBE_LDFLAGS)" $(MINIKUBE_TEST_FILES)

# Run the gotest, while recording JSON report and coverage
out/unittest.json: $(SOURCE_FILES) $(GOTEST_FILES)
	$(if $(quiet),@echo "  TEST     $@")
	$(Q)go test -tags "$(MINIKUBE_BUILD_TAGS)" -ldflags="$(MINIKUBE_LDFLAGS)" $(MINIKUBE_TEST_FILES) \
	-coverprofile=out/coverage.out -json > out/unittest.json
out/coverage.out: out/unittest.json

# Generate go test report (from gotest) as a HTML page
out/unittest.html: out/unittest.json
	$(if $(quiet),@echo "  REPORT   $@")
	$(Q)go-test-report < $< -o $@

# Generate go coverage report (from gotest) as a HTML page
out/coverage.html: out/coverage.out
	$(if $(quiet),@echo "  COVER    $@")
	$(Q)go tool cover -html=$< -o $@

.PHONY: extract
extract: ## extract internationalization words for translations
	go run cmd/extract/extract.go

.PHONY: cross
cross: minikube-linux-amd64 minikube-darwin-amd64 minikube-windows-amd64.exe ## Build minikube for all platform

.PHONY: exotic
exotic: out/minikube-linux-arm out/minikube-linux-arm64 out/minikube-linux-ppc64le out/minikube-linux-s390x ## Build minikube for non-amd64 linux

.PHONY: retro
retro: out/minikube-linux-386 out/minikube-linux-armv6 ## Build minikube for legacy 32-bit linux

.PHONY: windows
windows: minikube-windows-amd64.exe ## Build minikube for Windows 64bit

.PHONY: darwin
darwin: minikube-darwin-amd64 ## Build minikube for Darwin 64bit

.PHONY: linux
linux: minikube-linux-amd64 ## Build minikube for Linux 64bit

.PHONY: e2e-cross
e2e-cross: e2e-linux-amd64 e2e-linux-arm64 e2e-darwin-amd64 e2e-darwin-arm64 e2e-windows-amd64.exe ## End-to-end cross test

.PHONY: checksum
checksum: ## Generate checksums
	for f in out/minikube-amd64.iso out/minikube-arm64.iso out/minikube-linux-amd64 out/minikube-linux-arm \
		 out/minikube-linux-arm64 out/minikube-linux-ppc64le out/minikube-linux-s390x \
		 out/minikube-darwin-amd64 out/minikube-darwin-arm64 out/minikube-windows-amd64.exe \
		 out/docker-machine-driver-kvm2 out/docker-machine-driver-kvm2-amd64 out/docker-machine-driver-kvm2-arm64 \
		 out/docker-machine-driver-hyperkit; do \
		if [ -f "$${f}" ]; then \
			openssl sha256 "$${f}" | awk '{print $$2}' > "$${f}.sha256" ; \
		fi ; \
	done

.PHONY: clean
clean: ## Clean build
	rm -rf $(BUILD_DIR)
	rm -f pkg/minikube/assets/assets.go
	rm -f pkg/minikube/translate/translations.go
	rm -rf ./vendor
	rm -rf /tmp/tmp.*.minikube_*

.PHONY: gendocs
gendocs: out/docs/minikube.md  ## Generate documentation

.PHONY: fmt
fmt: ## Run go fmt and modify files in place
	@gofmt -s -w $(SOURCE_DIRS)

.PHONY: gofmt
gofmt: ## Run go fmt and list the files differs from gofmt's
	@gofmt -s -l $(SOURCE_DIRS)
	@test -z "`gofmt -s -l $(SOURCE_DIRS)`"

.PHONY: vet
vet: ## Run go vet
	@go vet $(SOURCE_PACKAGES)

.PHONY: imports
imports: ## Run goimports and modify files in place
	@goimports -w $(SOURCE_DIRS)

.PHONY: goimports
goimports: ## Run goimports and list the files differs from goimport's
	@goimports -l $(SOURCE_DIRS)
	@test -z "`goimports -l $(SOURCE_DIRS)`"

.PHONY: golint
golint: ## Run golint
	@golint -set_exit_status $(SOURCE_PACKAGES)

.PHONY: gocyclo
gocyclo: ## Run gocyclo (calculates cyclomatic complexities)
	@gocyclo -over 15 `find $(SOURCE_DIRS) -type f -name "*.go"`

out/linters/golangci-lint-$(GOLINT_VERSION):
	mkdir -p out/linters
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b out/linters $(GOLINT_VERSION)
	mv out/linters/golangci-lint out/linters/golangci-lint-$(GOLINT_VERSION)

# this one is meant for local use
.PHONY: lint
ifeq ($(MINIKUBE_BUILD_IN_DOCKER),y)
lint:
	docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:$(GOLINT_VERSION) \
	golangci-lint run ${GOLINT_OPTIONS} --skip-dirs "cmd/drivers/kvm|cmd/drivers/hyperkit|pkg/drivers/kvm|pkg/drivers/hyperkit" ./...
else
lint: out/linters/golangci-lint-$(GOLINT_VERSION) ## Run lint
	./out/linters/golangci-lint-$(GOLINT_VERSION) run ${GOLINT_OPTIONS} ./...
endif

# lint-ci is slower version of lint and is meant to be used in ci (travis) to avoid out of memory leaks.
.PHONY: lint-ci
lint-ci: out/linters/golangci-lint-$(GOLINT_VERSION) ## Run lint-ci
	GOGC=${GOLINT_GOGC} ./out/linters/golangci-lint-$(GOLINT_VERSION) run \
	--concurrency ${GOLINT_JOBS} ${GOLINT_OPTIONS} ./...

.PHONY: reportcard
reportcard: ## Run goreportcard for minikube
	goreportcard-cli -v
	# "disabling misspell on large repo..."
	-misspell -error $(SOURCE_DIRS)

.PHONY: mdlint
mdlint:
	@$(MARKDOWNLINT) $(MINIKUBE_MARKDOWN_FILES)

.PHONY: verify-iso
verify-iso: # Make sure the current ISO exists in the expected bucket
	gsutil stat gs://$(ISO_BUCKET)/minikube-$(ISO_VERSION)-amd64.iso
	gsutil stat gs://$(ISO_BUCKET)/minikube-$(ISO_VERSION)-arm64.iso

out/docs/minikube.md: $(shell find "cmd") $(shell find "pkg/minikube/constants")
	go run -ldflags="$(MINIKUBE_LDFLAGS)" -tags gendocs hack/help_text/gen_help_text.go

.PHONY: debs ## Build all deb packages
debs: out/minikube_$(DEB_VERSION)-$(DEB_REVISION)_amd64.deb \
	  out/minikube_$(DEB_VERSION)-$(DEB_REVISION)_arm64.deb \
	  out/docker-machine-driver-kvm2_$(DEB_VERSION).deb \
	  out/docker-machine-driver-kvm2_$(DEB_VERSION)-$(DEB_REVISION)_amd64.deb \
	  out/docker-machine-driver-kvm2_$(DEB_VERSION)-$(DEB_REVISION)_arm64.deb

.PHONY: deb_version
deb_version:
	@echo $(DEB_VERSION)-$(DEB_REVISION)

.PHONY: deb_version_base
deb_version_base:
	@echo $(DEB_VERSION)

out/minikube_$(DEB_VERSION).deb: out/minikube_$(DEB_VERSION)-$(DEB_REVISION)_amd64.deb
	cp $< $@

out/minikube_$(DEB_VERSION)-$(DEB_REVISION)_%.deb: out/minikube-linux-%
	$(eval DEB_PACKAGING_DIRECTORY_$*=$(shell mktemp -d --suffix ".minikube_$(DEB_VERSION)-$*-deb"))
	cp -r installers/linux/deb/minikube_deb_template/* $(DEB_PACKAGING_DIRECTORY_$*)/
	chmod 0755 $(DEB_PACKAGING_DIRECTORY_$*)/DEBIAN
	sed -E -i 's/--VERSION--/'$(DEB_VERSION)'/g' $(DEB_PACKAGING_DIRECTORY_$*)/DEBIAN/control
	sed -E -i 's/--REVISION--/'$(DEB_REVISION)'/g' $(DEB_PACKAGING_DIRECTORY_$*)/DEBIAN/control
	sed -E -i 's/--ARCH--/'$*'/g' $(DEB_PACKAGING_DIRECTORY_$*)/DEBIAN/control

	if [ "$*" = "amd64" ]; then \
	    sed -E -i 's/--RECOMMENDS--/virtualbox/' $(DEB_PACKAGING_DIRECTORY_$*)/DEBIAN/control; \
	else \
	    sed -E -i '/Recommends: --RECOMMENDS--/d' $(DEB_PACKAGING_DIRECTORY_$*)/DEBIAN/control; \
	fi

	mkdir -p $(DEB_PACKAGING_DIRECTORY_$*)/usr/bin
	cp $< $(DEB_PACKAGING_DIRECTORY_$*)/usr/bin/minikube
	fakeroot dpkg-deb --build $(DEB_PACKAGING_DIRECTORY_$*) $@
	rm -rf $(DEB_PACKAGING_DIRECTORY_$*)

rpm_version:
	@echo $(RPM_VERSION)-$(RPM_REVISION)

out/minikube-$(RPM_VERSION).rpm: out/minikube-$(RPM_VERSION)-$(RPM_REVISION).x86_64.rpm
	cp $< $@

out/minikube-$(RPM_VERSION)-0.%.rpm: out/minikube-linux-%
	$(eval RPM_PACKAGING_DIRECTORY_$*=$(shell mktemp -d --suffix ".minikube_$(RPM_VERSION)-$*-rpm"))
	cp -r installers/linux/rpm/minikube_rpm_template/* $(RPM_PACKAGING_DIRECTORY_$*)/
	sed -E -i 's/--VERSION--/'$(RPM_VERSION)'/g' $(RPM_PACKAGING_DIRECTORY_$*)/minikube.spec
	sed -E -i 's/--REVISION--/'$(RPM_REVISION)'/g' $(RPM_PACKAGING_DIRECTORY_$*)/minikube.spec
	sed -E -i 's|--OUT--|'$(PWD)/out'|g' $(RPM_PACKAGING_DIRECTORY_$*)/minikube.spec
	rpmbuild -bb -D "_rpmdir $(PWD)/out" --target $* \
		 $(RPM_PACKAGING_DIRECTORY_$*)/minikube.spec
	@mv out/$*/minikube-$(RPM_VERSION)-$(RPM_REVISION).$*.rpm out/ && rmdir out/$*
	rm -rf $(RPM_PACKAGING_DIRECTORY_$*)

.PHONY: apt
apt: out/Release ## Generate apt package file

out/Release: out/minikube_$(DEB_VERSION).deb
	( cd out && apt-ftparchive packages . ) | gzip -c > out/Packages.gz
	( cd out && apt-ftparchive release . ) > out/Release

.PHONY: yum
yum: out/repodata/repomd.xml

out/repodata/repomd.xml: out/minikube-$(RPM_VERSION).rpm
	createrepo --simple-md-filenames --no-database \
	-u "$(MINIKUBE_RELEASES_URL)/$(VERSION)/" out

.SECONDEXPANSION:
TAR_TARGETS_linux-amd64   := out/minikube-linux-amd64 out/docker-machine-driver-kvm2
TAR_TARGETS_linux-arm64   := out/minikube-linux-arm64 #out/docker-machine-driver-kvm2
TAR_TARGETS_darwin-amd64  := out/minikube-darwin-amd64 out/docker-machine-driver-hyperkit
TAR_TARGETS_darwin-arm64  := out/minikube-darwin-arm64 #out/docker-machine-driver-hyperkit
TAR_TARGETS_windows-amd64 := out/minikube-windows-amd64.exe
out/minikube-%.tar.gz: $$(TAR_TARGETS_$$*)
	$(if $(quiet),@echo "  TAR      $@")
	$(Q)tar -cvzf $@ $^

.PHONY: cross-tars
cross-tars: out/minikube-linux-amd64.tar.gz out/minikube-windows-amd64.tar.gz out/minikube-darwin-amd64.tar.gz ## Cross-compile minikube
	-cd out && $(SHA512SUM) *.tar.gz > SHA512SUM

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

out/docker-machine-driver-hyperkit:
ifeq ($(MINIKUBE_BUILD_IN_DOCKER),y)
	docker run --rm -e GOCACHE=/app/.cache -e IN_DOCKER=1 \
		--user $(shell id -u):$(shell id -g) -w /app \
		-v $(PWD):/app -v $(GOPATH):/go --init --entrypoint "" \
		$(HYPERKIT_BUILD_IMAGE) /bin/bash -c 'CC=o64-clang CXX=o64-clang++ /usr/bin/make $@'
else
	$(if $(quiet),@echo "  GO       $@")
	$(Q)GOOS=darwin CGO_ENABLED=1 go build \
		-ldflags="$(HYPERKIT_LDFLAGS)"   \
		-o $@ k8s.io/minikube/cmd/drivers/hyperkit
endif

hyperkit_in_docker:
	rm -f out/docker-machine-driver-hyperkit
	$(MAKE) MINIKUBE_BUILD_IN_DOCKER=y out/docker-machine-driver-hyperkit

.PHONY: install-hyperkit-driver
install-hyperkit-driver: out/docker-machine-driver-hyperkit ## Install hyperkit to local machine
	mkdir -p $(HOME)/bin
	sudo cp out/docker-machine-driver-hyperkit $(HOME)/bin/docker-machine-driver-hyperkit
	sudo chown root:wheel $(HOME)/bin/docker-machine-driver-hyperkit
	sudo chmod u+s $(HOME)/bin/docker-machine-driver-hyperkit

.PHONY: release-hyperkit-driver
release-hyperkit-driver: install-hyperkit-driver checksum ## Copy hyperkit using gsutil
	gsutil cp $(GOBIN)/docker-machine-driver-hyperkit gs://minikube/drivers/hyperkit/$(VERSION)/
	gsutil cp $(GOBIN)/docker-machine-driver-hyperkit.sha256 gs://minikube/drivers/hyperkit/$(VERSION)/

.PHONY: build-and-push-hyperkit-build-image
build-and-push-hyperkit-build-image:
	test -d out/xcgo || git clone https://github.com/neilotoole/xcgo.git out/xcgo
	(cd out/xcgo && git restore . && git pull && \
	 sed -i'.bak' -e 's/ARG GO_VERSION.*/ARG GO_VERSION="go$(GO_VERSION)"/' Dockerfile && \
	 docker build -t gcr.io/k8s-minikube/xcgo:go$(GO_VERSION) .)
	docker push gcr.io/k8s-minikube/xcgo:go$(GO_VERSION)

.PHONY: check-release
check-release: ## Execute go test
	go test -timeout 42m -v ./deploy/minikube/release_sanity_test.go

buildroot-image: $(ISO_BUILD_IMAGE) # convenient alias to build the docker container
$(ISO_BUILD_IMAGE): deploy/iso/minikube-iso/Dockerfile
	docker build $(ISO_DOCKER_EXTRA_ARGS) -t $@ -f $< $(dir $<)
	@echo ""
	@echo "$(@) successfully built"

out/storage-provisioner: out/storage-provisioner-$(GOARCH)
	$(if $(quiet),@echo "  CP       $@")
	$(Q)cp $< $@

out/storage-provisioner-%: cmd/storage-provisioner/main.go pkg/storage/storage_provisioner.go
ifeq ($(MINIKUBE_BUILD_IN_DOCKER),y)
	$(call DOCKER,$(BUILD_IMAGE),/usr/bin/make $@)
else
	$(if $(quiet),@echo "  GO       $@")
	$(Q)CGO_ENABLED=0 GOOS=linux GOARCH=$* go build -o $@ -ldflags=$(PROVISIONER_LDFLAGS) cmd/storage-provisioner/main.go
endif

.PHONY: storage-provisioner-image
storage-provisioner-image: storage-provisioner-image-$(GOARCH) ## Build storage-provisioner docker image
	docker tag $(REGISTRY)/storage-provisioner-$(GOARCH):$(STORAGE_PROVISIONER_TAG) $(REGISTRY)/storage-provisioner:$(STORAGE_PROVISIONER_TAG)

storage-provisioner-image-%: out/storage-provisioner-%
	docker build -t $(REGISTRY)/storage-provisioner-$*:$(STORAGE_PROVISIONER_TAG) -f deploy/storage-provisioner/Dockerfile  --build-arg arch=$* .


.PHONY: docker-multi-arch-build
docker-multi-arch-build:
	# installs QEMU static binaries to allow docker multi-arch build, see: https://github.com/docker/setup-qemu-action
	docker run --rm --privileged tonistiigi/binfmt:latest --install all

KICBASE_ARCH ?= linux/amd64,linux/arm64,linux/s390x,linux/arm,linux/ppc64le
KICBASE_IMAGE_GCR ?= $(REGISTRY)/kicbase:$(KIC_VERSION)
KICBASE_IMAGE_HUB ?= kicbase/stable:$(KIC_VERSION)
KICBASE_IMAGE_REGISTRIES ?= $(KICBASE_IMAGE_GCR) $(KICBASE_IMAGE_HUB)

.PHONY: build-and-upload-cri-dockerd-binaries
build-and-upload-cri-dockerd-binaries:
	(cd hack/update/cri_dockerd_version && \
	 ./build_and_upload_cri_dockerd_binaries.sh $(KICBASE_ARCH))

.PHONY: local-kicbase
local-kicbase: ## Builds the kicbase image and tags it local/kicbase:latest and local/kicbase:$(KIC_VERSION)-$(COMMIT_SHORT)
	touch deploy/kicbase/CHANGELOG
	docker build -f ./deploy/kicbase/Dockerfile -t local/kicbase:$(KIC_VERSION) --build-arg VERSION_JSON=$(VERSION_JSON) --build-arg COMMIT_SHA=${VERSION}-$(COMMIT_NOQUOTES) --cache-from $(KICBASE_IMAGE_GCR) .
	docker tag local/kicbase:$(KIC_VERSION) local/kicbase:latest
	docker tag local/kicbase:$(KIC_VERSION) local/kicbase:$(KIC_VERSION)-$(COMMIT_SHORT)


.PHONY: local-kicbase-debug
local-kicbase-debug: local-kicbase ## Builds a local kicbase image and switches source code to point to it
	$(SED) 's|Version = .*|Version = \"$(KIC_VERSION)-$(COMMIT_SHORT)\"|;s|baseImageSHA = .*|baseImageSHA = \"\"|;s|gcrRepo = .*|gcrRepo = \"local/kicbase\"|;s|dockerhubRepo = .*|dockerhubRepo = \"local/kicbase\"|' pkg/drivers/kic/types.go

.PHONY: build-kic-base-image
build-kic-base-image: docker-multi-arch-build ## Build multi-arch local/kicbase:latest
	docker buildx build -f ./deploy/kicbase/Dockerfile --platform $(KICBASE_ARCH) $(addprefix -t ,$(KICBASE_IMAGE_REGISTRIES)) --build-arg VERSION_JSON=$(VERSION_JSON) --build-arg COMMIT_SHA=${VERSION}-$(COMMIT_NOQUOTES) .

.PHONY: push-kic-base-image
push-kic-base-image: docker-multi-arch-build ## Push multi-arch local/kicbase:latest to all remote registries
ifdef AUTOPUSH
	docker login gcr.io/k8s-minikube
	docker login docker.pkg.github.com
	docker login
endif
	$(foreach REG,$(KICBASE_IMAGE_REGISTRIES), \
		@docker pull $(REG) && echo "Image already exist in registry" && exit 1 || echo "Image doesn't exist in registry";)
ifndef CIBUILD
	$(call user_confirm, 'Are you sure you want to push $(KICBASE_IMAGE_REGISTRIES) ?')
endif
	./deploy/kicbase/build_auto_pause.sh $(KICBASE_ARCH)
	docker buildx build -f ./deploy/kicbase/Dockerfile --platform $(KICBASE_ARCH) $(addprefix -t ,$(KICBASE_IMAGE_REGISTRIES)) --push --build-arg VERSION_JSON=$(VERSION_JSON) --build-arg COMMIT_SHA=${VERSION}-$(COMMIT_NOQUOTES) --build-arg PREBUILT_AUTO_PAUSE=true .

out/preload-tool:
	go build -ldflags="$(MINIKUBE_LDFLAGS)" -o $@ ./hack/preload-images/*.go

.PHONY: upload-preloaded-images-tar
upload-preloaded-images-tar: out/minikube out/preload-tool ## Upload the preloaded images for oldest supported, newest supported, and default kubernetes versions to GCS.
	out/preload-tool

.PHONY: generate-preloaded-images-tar
generate-preloaded-images-tar: out/minikube out/preload-tool ## Generates the preloaded images for oldest supported, newest supported, and default kubernetes versions
	out/preload-tool --no-upload

ALL_ARCH = amd64 arm arm64 ppc64le s390x
IMAGE = $(REGISTRY)/storage-provisioner
TAG = $(STORAGE_PROVISIONER_TAG)

.PHONY: push-storage-provisioner-manifest
push-storage-provisioner-manifest: $(shell echo $(ALL_ARCH) | sed -e "s~[^ ]*~storage\-provisioner\-image\-&~g") ## Push multi-arch storage-provisioner image
ifndef CIBUILD
	docker login gcr.io/k8s-minikube
endif
	set -x; for arch in $(ALL_ARCH); do docker push ${IMAGE}-$${arch}:${TAG}; done
	docker manifest create --amend $(IMAGE):$(TAG) $(shell echo $(ALL_ARCH) | sed -e "s~[^ ]*~$(IMAGE)\-&:$(TAG)~g")
	set -x; for arch in $(ALL_ARCH); do docker manifest annotate --arch $${arch} ${IMAGE}:${TAG} ${IMAGE}-$${arch}:${TAG}; done
	docker manifest push $(STORAGE_PROVISIONER_MANIFEST)

.PHONY: push-docker
push-docker: # Push docker image base on to IMAGE variable (used internally by other targets)
	@docker pull $(IMAGE) && echo "Image already exist in registry" && exit 1 || echo "Image doesn't exist in registry"
ifndef AUTOPUSH
	$(call user_confirm, 'Are you sure you want to push $(IMAGE) ?')
endif
	docker push $(IMAGE)

.PHONY: out/gvisor-addon
out/gvisor-addon: ## Build gvisor addon
	$(if $(quiet),@echo "  GO       $@")
	$(Q)GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $@ cmd/gvisor/gvisor.go

.PHONY: gvisor-addon-image
gvisor-addon-image:
	docker build -t $(REGISTRY)/gvisor-addon:$(GVISOR_TAG) -f deploy/gvisor/Dockerfile .

.PHONY: push-gvisor-addon-image
push-gvisor-addon-image: docker-multi-arch-build
	docker login gcr.io/k8s-minikube
	docker buildx create --name multiarch --bootstrap
	docker buildx build --push --builder multiarch --platform linux/amd64,linux/arm64 -t $(REGISTRY)/gvisor-addon:$(GVISOR_TAG) -t $(REGISTRY)/gvisor-addon:latest -f deploy/gvisor/Dockerfile .
	docker buildx rm multiarch

.PHONY: release-iso
release-iso: minikube-iso-aarch64 minikube-iso-x86_64 checksum  ## Build and release .iso files
	gsutil cp out/minikube-amd64.iso gs://$(ISO_BUCKET)/minikube-$(ISO_VERSION)-amd64.iso
	gsutil cp out/minikube-amd64.iso.sha256 gs://$(ISO_BUCKET)/minikube-$(ISO_VERSION)-amd64.iso.sha256
	gsutil cp out/minikube-arm64.iso gs://$(ISO_BUCKET)/minikube-$(ISO_VERSION)-arm64.iso
	gsutil cp out/minikube-arm64.iso.sha256 gs://$(ISO_BUCKET)/minikube-$(ISO_VERSION)-arm64.iso.sha256

.PHONY: release-minikube
release-minikube: out/minikube checksum ## Minikube release
	gsutil cp out/minikube-$(GOOS)-$(GOARCH) $(MINIKUBE_UPLOAD_LOCATION)/$(MINIKUBE_VERSION)/minikube-$(GOOS)-$(GOARCH)
	gsutil cp out/minikube-$(GOOS)-$(GOARCH).sha256 $(MINIKUBE_UPLOAD_LOCATION)/$(MINIKUBE_VERSION)/minikube-$(GOOS)-$(GOARCH).sha256

.PHONY: release-notes
release-notes:
	hack/release_notes.sh

.PHONY: update-leaderboard
update-leaderboard:
	hack/update_contributions.sh

.PHONY: update-yearly-leaderboard
update-yearly-leaderboard:
	hack/yearly-leaderboard.sh

out/docker-machine-driver-kvm2: out/docker-machine-driver-kvm2-$(GOARCH)
	$(if $(quiet),@echo "  CP       $@")
	$(Q)cp $< $@

out/docker-machine-driver-kvm2-x86_64: out/docker-machine-driver-kvm2-amd64
	$(if $(quiet),@echo "  CP       $@")
	$(Q)cp $< $@

out/docker-machine-driver-kvm2-aarch64: out/docker-machine-driver-kvm2-arm64
	$(if $(quiet),@echo "  CP       $@")
	$(Q)cp $< $@


out/docker-machine-driver-kvm2_$(DEB_VERSION).deb: out/docker-machine-driver-kvm2_$(DEB_VERSION)-0_amd64.deb
	cp $< $@

out/docker-machine-driver-kvm2_$(DEB_VERSION)-0_%.deb: out/docker-machine-driver-kvm2-%
	cp -r installers/linux/deb/kvm2_deb_template out/docker-machine-driver-kvm2_$(DEB_VERSION)
	chmod 0755 out/docker-machine-driver-kvm2_$(DEB_VERSION)/DEBIAN
	sed -E -i -e 's/--VERSION--/$(DEB_VERSION)/g' out/docker-machine-driver-kvm2_$(DEB_VERSION)/DEBIAN/control
	sed -E -i -e 's/--ARCH--/'$*'/g' out/docker-machine-driver-kvm2_$(DEB_VERSION)/DEBIAN/control
	mkdir -p out/docker-machine-driver-kvm2_$(DEB_VERSION)/usr/bin
	cp $< out/docker-machine-driver-kvm2_$(DEB_VERSION)/usr/bin/docker-machine-driver-kvm2
	fakeroot dpkg-deb --build out/docker-machine-driver-kvm2_$(DEB_VERSION) $@
	rm -rf out/docker-machine-driver-kvm2_$(DEB_VERSION)

out/docker-machine-driver-kvm2-$(RPM_VERSION).rpm: out/docker-machine-driver-kvm2-$(RPM_VERSION)-0.x86_64.rpm
	cp $< $@

out/docker-machine-driver-kvm2_$(RPM_VERSION).amd64.rpm: out/docker-machine-driver-kvm2-$(RPM_VERSION)-0.x86_64.rpm
	cp $< $@

out/docker-machine-driver-kvm2_$(RPM_VERSION).arm64.rpm: out/docker-machine-driver-kvm2-$(RPM_VERSION)-0.aarch64.rpm
	cp $< $@

out/docker-machine-driver-kvm2-$(RPM_VERSION)-0.%.rpm: out/docker-machine-driver-kvm2-%
	cp -r installers/linux/rpm/kvm2_rpm_template out/docker-machine-driver-kvm2-$(RPM_VERSION)
	sed -E -i -e 's/--VERSION--/'$(RPM_VERSION)'/g' out/docker-machine-driver-kvm2-$(RPM_VERSION)/docker-machine-driver-kvm2.spec
	sed -E -i -e 's|--OUT--|'$(PWD)/out'|g' out/docker-machine-driver-kvm2-$(RPM_VERSION)/docker-machine-driver-kvm2.spec
	rpmbuild -bb -D "_rpmdir $(PWD)/out" --target $* \
		out/docker-machine-driver-kvm2-$(RPM_VERSION)/docker-machine-driver-kvm2.spec
	@mv out/$*/docker-machine-driver-kvm2-$(RPM_VERSION)-0.$*.rpm out/ && rmdir out/$*
	rm -rf out/docker-machine-driver-kvm2-$(RPM_VERSION)

.PHONY: kvm-image-amd64
kvm-image-amd64: installers/linux/kvm/Dockerfile.amd64  ## Convenient alias to build the docker container
	docker build --build-arg "GO_VERSION=$(GO_VERSION)" -t $(KVM_BUILD_IMAGE_AMD64) -f $< $(dir $<)
	@echo ""
	@echo "$(@) successfully built"

.PHONY: kvm-image-arm64
kvm-image-arm64: installers/linux/kvm/Dockerfile.arm64 docker-multi-arch-build  ## Convenient alias to build the docker container
	docker buildx build --platform linux/arm64 --build-arg "GO_VERSION=$(GO_VERSION)" -t $(KVM_BUILD_IMAGE_ARM64) -f $< $(dir $<)
	@echo ""
	@echo "$(@) successfully built"

kvm_in_docker:
	docker image inspect -f '{{.Id}} {{.RepoTags}}' $(KVM_BUILD_IMAGE_AMD64) || $(MAKE) kvm-image-amd64
	rm -f out/docker-machine-driver-kvm2
	$(call DOCKER,$(KVM_BUILD_IMAGE_AMD64),/usr/bin/make out/docker-machine-driver-kvm2 COMMIT=$(COMMIT))

.PHONY: install-kvm-driver
install-kvm-driver: out/docker-machine-driver-kvm2  ## Install KVM Driver
	mkdir -p $(GOBIN)
	cp out/docker-machine-driver-kvm2 $(GOBIN)/docker-machine-driver-kvm2


out/docker-machine-driver-kvm2-arm64:
ifeq ($(MINIKUBE_BUILD_IN_DOCKER),y)
	docker image inspect -f '{{.Id}} {{.RepoTags}}' $(KVM_BUILD_IMAGE_ARM64) || $(MAKE) kvm-image-arm64
	$(call DOCKER,$(KVM_BUILD_IMAGE_ARM64),/usr/bin/make $@ COMMIT=$(COMMIT))
else
	$(if $(quiet),@echo "  GO       $@")
	$(Q)GOARCH=arm64 \
	go build \
		-buildvcs=false \
		-installsuffix "static" \
		-ldflags="$(KVM2_LDFLAGS)" \
		-tags "libvirt_without_lxc" \
		-o $@ \
		k8s.io/minikube/cmd/drivers/kvm
endif
	chmod +X $@

out/docker-machine-driver-kvm2-%:
ifeq ($(MINIKUBE_BUILD_IN_DOCKER),y)
	docker image inspect -f '{{.Id}} {{.RepoTags}}' $(KVM_BUILD_IMAGE_AMD64) || $(MAKE) kvm-image-amd64
	$(call DOCKER,$(KVM_BUILD_IMAGE_AMD64),/usr/bin/make $@ COMMIT=$(COMMIT))
else
	$(if $(quiet),@echo "  GO       $@")
	$(Q)GOARCH=$* \
	go build \
	        -buildvcs=false \
		-installsuffix "static" \
		-ldflags="$(KVM2_LDFLAGS)" \
		-tags "libvirt_without_lxc" \
		-o $@ \
		k8s.io/minikube/cmd/drivers/kvm
endif
	chmod +X $@


site/themes/docsy/assets/vendor/bootstrap/package.js: ## update the website docsy theme git submodule 
	git submodule update -f --init

.PHONY: out/hugo/hugo
out/hugo/hugo:
	mkdir -p out
	(cd site/themes/docsy && npm install)
	test -d out/hugo || git clone https://github.com/gohugoio/hugo.git out/hugo
	(cd out/hugo && git fetch origin && git checkout $(HUGO_VERSION) && go build --tags extended)

.PHONY: site
site: site/themes/docsy/assets/vendor/bootstrap/package.js out/hugo/hugo ## Serve the documentation site to localhost
	(cd site && ../out/hugo/hugo serve \
	  --disableFastRender \
	  --navigateToChanged \
	  --ignoreCache \
	  --buildFuture)

.PHONY: out/mkcmp
out/mkcmp:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $@ cmd/performance/mkcmp/main.go


# auto pause binary to be used for ISO
deploy/iso/minikube-iso/board/minikube/%/rootfs-overlay/usr/bin/auto-pause: $(SOURCE_FILES) $(ASSET_FILES)
	@if [ "$*" != "x86_64" ] && [ "$*" != "aarch64" ]; then echo "Please enter a valid architecture. Choices are x86_64 and aarch64."; exit 1; fi
	GOOS=linux GOARCH=$(subst x86_64,amd64,$(subst aarch64,arm64,$*)) go build -o $@ cmd/auto-pause/auto-pause.go


.PHONY: deploy/addons/auto-pause/auto-pause-hook
deploy/addons/auto-pause/auto-pause-hook: ## Build auto-pause hook addon
	$(if $(quiet),@echo "  GO       $@")
	$(Q)GOOS=linux CGO_ENABLED=0 go build -a --ldflags '-extldflags "-static"' -tags netgo -installsuffix netgo -o $@ cmd/auto-pause/auto-pause-hook/main.go cmd/auto-pause/auto-pause-hook/config.go cmd/auto-pause/auto-pause-hook/certs.go

.PHONY: auto-pause-hook-image
auto-pause-hook-image: deploy/addons/auto-pause/auto-pause-hook ## Build docker image for auto-pause hook
	docker build -t $(REGISTRY)/auto-pause-hook:$(AUTOPAUSE_HOOK_TAG) ./deploy/addons/auto-pause

.PHONY: push-auto-pause-hook-image
push-auto-pause-hook-image: docker-multi-arch-build
	docker login gcr.io/k8s-minikube
	docker buildx create --name multiarch --bootstrap
	docker buildx build --push --builder multiarch --platform $(KICBASE_ARCH) -t $(REGISTRY)/auto-pause-hook:$(AUTOPAUSE_HOOK_TAG) -f ./deploy/addons/auto-pause/Dockerfile .
	docker buildx rm multiarch

.PHONY: push-prow-test-image
push-prow-test-image: docker-multi-arch-build
	docker login gcr.io/k8s-minikube
	docker buildx create --name multiarch --bootstrap
	docker buildx build --push --builder multiarch --build-arg "GO_VERSION=$(GO_VERSION)" --platform linux/amd64,linux/arm64 -t $(REGISTRY)/prow-test:$(PROW_TEST_TAG) -t $(REGISTRY)/prow-test:latest ./deploy/prow
	docker buildx rm multiarch

.PHONY: out/performance-bot
out/performance-bot:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $@ cmd/performance/pr-bot/bot.go

.PHONY: out/metrics-collector
out/metrics-collector:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $@ hack/metrics/*.go


.PHONY: compare
compare: out/mkcmp out/minikube
	mv out/minikube out/$(CURRENT_GIT_BRANCH).minikube
	git checkout master
	make out/minikube
	mv out/minikube out/master.minikube
	git checkout $(CURRENT_GIT_BRANCH)
	out/mkcmp out/master.minikube out/$(CURRENT_GIT_BRANCH).minikube
	

.PHONY: help
help:
	@printf "\033[1mAvailable targets for minikube ${VERSION}\033[21m\n"
	@printf "\033[1m--------------------------------------\033[21m\n"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'



.PHONY: update-golang-version
update-golang-version:
	(cd hack/update/golang_version && \
	 go run update_golang_version.go)

.PHONY: update-kubernetes-version
update-kubernetes-version:
	@(cd hack/update/kubernetes_version && \
	 go run update_kubernetes_version.go)

.PHONY: update-golint-version
update-golint-version:
	(cd hack/update/golint_version && \
	 go run update_golint_version.go)

.PHONY: update-preload-version
update-preload-version:
	(cd hack/update/preload_version && \
	 go run update_preload_version.go)

.PHONY: update-kubeadm-constants
update-kubeadm-constants:
	(cd hack/update/kubeadm_constants && \
	 go run update_kubeadm_constants.go)
	gofmt -w pkg/minikube/constants/constants_kubeadm_images.go

.PHONY: stress
stress: ## run the stress tests
	go test -test.v -test.timeout=2h ./test/stress -loops=10 | tee "./out/testout_$(COMMIT_SHORT).txt"

.PHONY: cpu-benchmark-idle
cpu-benchmark-idle: ## run the cpu usage 5 minutes idle benchmark
	./hack/benchmark/cpu_usage/idle_only/benchmark_local_k8s.sh

.PHONY: cpu-benchmark-autopause
cpu-benchmark-autopause: ## run the cpu usage auto-pause benchmark
	./hack/benchmark/cpu_usage/auto_pause/benchmark_local_k8s.sh

.PHONY: time-to-k8s-benchmark
time-to-k8s-benchmark:
	./hack/benchmark/time-to-k8s/time-to-k8s.sh

.PHONY: update-gopogh-version
update-gopogh-version: ## update gopogh version
	(cd hack/update/gopogh_version && \
	 go run update_gopogh_version.go)

.PHONY: update-gotestsum-version
update-gotestsum-version:
	(cd hack/update/gotestsum_version && \
	 go run update_gotestsum_version.go)

.PHONY: update-gh-version
update-gh-version:
	(cd hack/update/gh_version && \
	 go run update_gh_version.go)

.PHONY: update-docsy-version
update-docsy-version:
	@(cd hack/update/docsy_version && \
	  go run update_docsy_version.go)

.PHONY: update-hugo-version
update-hugo-version:
	(cd hack/update/hugo_version && \
	 go run update_hugo_version.go)

.PHONY: update-cloud-spanner-emulator-version
update-cloud-spanner-emulator-version:
	(cd hack/update/cloud_spanner_emulator_version && \
	 go run update_cloud_spanner_emulator_version.go)

.PHONY: update-containerd-version
update-containerd-version:
	(cd hack/update/containerd_version && \
	 go run update_containerd_version.go)

.PHONY: update-buildkit-version
update-buildkit-version:
	(cd hack/update/buildkit_version && \
	 go run update_buildkit_version.go)

.PHONY: update-cri-o-version
update-cri-o-version:
	(cd hack/update/cri-o_version && \
	 go run update_cri-o_version.go)

.PHONY: update-crun-version
update-crun-version:
	(cd hack/update/crun_version && \
	 go run update_crun_version.go)

.PHONY: update-metrics-server-version
update-metrics-server-version:
	(cd hack/update/metrics_server_version && \
	 go run update_metrics_server_version.go)

.PHONY: update-runc-version
update-runc-version:
	(cd hack/update/runc_version && \
	 go run update_runc_version.go)

.PHONY: update-docker-version
update-docker-version:
	(cd hack/update/docker_version && \
	 go run update_docker_version.go)

.PHONY: update-ubuntu-version
update-ubuntu-version:
	(cd hack/update/ubuntu_version && \
	 go run update_ubuntu_version.go)

.PHONY: update-cni-plugins-version
update-cni-plugins-version:
	(cd hack/update/cni_plugins_version && \
	 go run update_cni_plugins_version.go)

.PHONY: update-gcp-auth-version
update-gcp-auth-version:
	(cd hack/update/gcp_auth_version && \
	 go run update_gcp_auth_version.go)

.PHONY: update-kubernetes-versions-list
update-kubernetes-versions-list:
	(cd hack/update/kubernetes_versions_list && \
	 go run update_kubernetes_versions_list.go)

.PHONY: update-ingress-version
update-ingress-version:
	(cd hack/update/ingress_version && \
	 go run update_ingress_version.go)

.PHONY: update-flannel-version
update-flannel-version:
	(cd hack/update/flannel_version && \
	 go run update_flannel_version.go)

.PHONY: update-inspektor-gadget-version
update-inspektor-gadget-version:
	(cd hack/update/inspektor_gadget_version && \
	 go run update_inspektor_gadget_version.go)

.PHONY: update-calico-version
update-calico-version:
	(cd hack/update/calico_version && \
	 go run update_calico_version.go)

.PHONY: update-cri-dockerd-version
update-cri-dockerd-version:
	(cd hack/update/cri_dockerd_version && \
	 go run update_cri_dockerd_version.go)

.PHONY: update-go-github-version
update-go-github-version:
	(cd hack/update/go_github_version && \
	 go run update_go_github_version.go)

.PHONY: update-docker-buildx-version
update-docker-buildx-version:
	(cd hack/update/docker_buildx_version && \
	 go run update_docker_buildx_version.go)

.PHONY: update-nerdctl-version
update-nerdctl-version:
	(cd hack/update/nerdctl_version && \
	 go run update_nerdctl_version.go)

.PHONY: update-crictl-version
update-crictl-version:
	(cd hack/update/crictl_version && \
	 go run update_crictl_version.go)

.PHONY: update-kindnetd-version
update-kindnetd-version:
	(cd hack/update/kindnetd_version && \
	 go run update_kindnetd_version.go)

.PHONY: update-istio-operator-version
update-istio-operator-version:
	(cd hack/update/istio_operator_version && \
	 go run update_istio_operator_version.go)

.PHONY: update-registry-version
update-registry-version:
	(cd hack/update/registry_version && \
	 go run update_registry_version.go)

.PHONY: update-volcano-version
update-volcano-version:
	(cd hack/update/volcano_version && \
	 go run update_volcano_version.go)

.PHONY: update-kong-version
update-kong-version:
	(cd hack/update/kong_version && \
	 go run update_kong_version.go)

.PHONY: update-kong-ingress-controller-version
update-kong-ingress-controller-version:
	(cd hack/update/kong_ingress_controller_version && \
	 go run update_kong_ingress_controller_version.go)

.PHONY: update-nvidia-device-plugin-version
update-nvidia-device-plugin-version:
	(cd hack/update/nvidia_device_plugin_version && \
	 go run update_nvidia_device_plugin_version.go)

.PHONY: update-amd-gpu-device-plugin-version
update-amd-gpu-device-plugin-version:
	(cd hack/update/amd_device_plugin_version && \
	 go run update_amd_device_plugin_version.go)

.PHONY: update-nerctld-version
update-nerdctld-version:
	(cd hack/update/nerdctld_version && \
	 go run update_nerdctld_version.go)

.PHONY: update-kubectl-version
update-kubectl-version:
	(cd hack/update/kubectl_version && \
	 go run update_kubectl_version.go)

.PHONY: update-site-node-version
update-site-node-version:
	(cd hack/update/site_node_version && \
	 go run update_site_node_version.go)

.PHONY: update-cilium-version
update-cilium-version:
	(cd hack/update/cilium_version && \
	 go run update_cilium_version.go)

.PHONY: update-yakd-version
update-yakd-version:
	(cd hack/update/yakd_version && \
	 go run update_yakd_version.go)

.PHONY: update-kube-registry-proxy-version
update-kube-registry-proxy-version:
	(cd hack/update/kube_registry_proxy_version && \
	 go run update_kube_registry_proxy_version.go)

.PHONY: get-dependency-verison
get-dependency-version:
	@(cd hack/update/get_version && \
	  go run get_version.go)

.PHONY: generate-licenses
generate-licenses:
	./hack/generate_licenses.sh

.PHONY: update-kube-vip-version
update-kube-vip-version:
	(cd hack/update/kube_vip_version && \
	 go run update_kube_vip_version.go)
