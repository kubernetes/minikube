################################################################################
#
# cri-dockerd
#
################################################################################

# As of 2023-01-19
CRI_DOCKERD_AARCH64_VER = 0.3.0
CRI_DOCKERD_AARCH64_REV = 0de30fc
CRI_DOCKERD_AARCH64_VERSION = 0de30fc57b659cf23b1212d6516e0cceab9c91d1
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
