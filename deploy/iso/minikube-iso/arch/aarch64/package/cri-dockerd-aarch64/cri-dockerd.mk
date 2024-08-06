################################################################################
#
# cri-dockerd
#
################################################################################

CRI_DOCKERD_AARCH64_VER = 0.3.15
CRI_DOCKERD_AARCH64_REV = c1c566e
CRI_DOCKERD_AARCH64_VERSION = c1c566e0cc84abe6972f0bf857ecd8fe306258d9
CRI_DOCKERD_AARCH64_SITE = https://github.com/Mirantis/cri-dockerd/archive
CRI_DOCKERD_AARCH64_SOURCE = $(CRI_DOCKERD_AARCH64_VERSION).tar.gz

CRI_DOCKERD_AARCH64_DEPENDENCIES = host-go
CRI_DOCKERD_AARCH64_GOPATH = $(@D)/_output
CRI_DOCKERD_AARCH64_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=0 \
	GO111MODULE=on \
	GOPATH="$(CRI_DOCKERD_AARCH64_GOPATH)" \
	PATH=$(CRI_DOCKERD_AARCH64_GOPATH)/bin:$(BR_PATH) \
	GOARCH=arm64

CRI_DOCKERD_AARCH64_COMPILE_SRC = $(CRI_DOCKERD_AARCH64_GOPATH)/src/github.com/Mirantis/cri-dockerd
CRI_DOCKERD_AARCH64_BUILDFLAGS = "-ldflags '-X github.com/Mirantis/cri-dockerd/version.Version=$(CRI_DOCKERD_AARCH64_VER) -X github.com/Mirantis/cri-dockerd/version.GitCommit=$(CRI_DOCKERD_AARCH64_REV)'"

define CRI_DOCKERD_AARCH64_POST_EXTRACT_WORKAROUNDS
	# Set -buildvcs=false to disable VCS stamping (fails in buildroot)
	sed -i 's|go build |go build -buildvcs=false |' -i $(@D)/packaging/static/Makefile
	# Use the GOARCH environment variable that we set
	sed -i 's|GOARCH=$(ARCH) go build|GOARCH=$(GOARCH) go build|' -i $(@D)/packaging/static/Makefile
endef

CRI_DOCKERD_AARCH64_POST_EXTRACT_HOOKS += CRI_DOCKERD_AARCH64_POST_EXTRACT_WORKAROUNDS

# If https://github.com/Mirantis/cri-dockerd/blob/master/packaging/Makefile changes, then this will almost certainly need to change
# This uses the static make target at the top level Makefile, since that builds everything, then picks out the arm64 binary
define CRI_DOCKERD_AARCH64_BUILD_CMDS
	$(CRI_DOCKERD_AARCH64_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) LDFLAGS=$(CRI_DOCKERD_AARCH64_BUILDFLAGS) GO_VERSION=$(GO_VERSION) -C $(@D) VERSION=$(CRI_DOCKERD_AARCH64_VER) REVISION=$(CRI_DOCKERD_AARCH64_REV) static
endef

define CRI_DOCKERD_AARCH64_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
		$(@D)/packaging/static/build/arm/cri-dockerd/cri-dockerd \
		$(TARGET_DIR)/usr/bin/cri-dockerd
endef

define CRI_DOCKERD_AARCH64_INSTALL_INIT_SYSTEMD
	$(INSTALL) -Dm644 \
		$(@D)/packaging/systemd/cri-docker.service \
		$(TARGET_DIR)/usr/lib/systemd/system/cri-docker.service
	$(INSTALL) -Dm644 \
		$(@D)/packaging/systemd/cri-docker.socket \
		$(TARGET_DIR)/usr/lib/systemd/system/cri-docker.socket
endef

$(eval $(generic-package))
