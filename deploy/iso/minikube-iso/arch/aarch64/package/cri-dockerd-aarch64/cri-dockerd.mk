################################################################################
#
# cri-dockerd
#
################################################################################

CRI_DOCKERD_AARCH64_VERSION = 0.4.1
CRI_DOCKERD_AARCH64_COMMIT = 55d6e1a1d6f2ee58949e13a0c66afe7d779ac942
CRI_DOCKERD_AARCH64_SITE = https://github.com/Mirantis/cri-dockerd/archive
CRI_DOCKERD_AARCH64_SOURCE = $(CRI_DOCKERD_AARCH64_COMMIT).tar.gz

CRI_DOCKERD_AARCH64_DEPENDENCIES = host-go
CRI_DOCKERD_AARCH64_GOPATH = $(@D)/_output
CRI_DOCKERD_AARCH64_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=0 \
	GO111MODULE=on \
	GOPATH="$(CRI_DOCKERD_AARCH64_GOPATH)" \
	PATH=$(CRI_DOCKERD_AARCH64_GOPATH)/bin:$(BR_PATH) \
	GOARCH=arm64 \
	GOPROXY="https://proxy.golang.org,direct" \
	GOSUMDB='sum.golang.org'\
	GOOS=linux

CRI_DOCKERD_AARCH64_COMPILE_SRC = $(CRI_DOCKERD_AARCH64_GOPATH)/src/github.com/Mirantis/cri-dockerd

define CRI_DOCKERD_AARCH64_POST_EXTRACT_WORKAROUNDS
	# Set -buildvcs=false to disable VCS stamping (fails in buildroot)
	sed -i 's|go build |go build -buildvcs=false |' -i $(@D)/Makefile
	# Use the GOARCH environment variable that we set
	sed -i 's|GOARCH=$$(ARCH) go build|GOARCH=$$(GOARCH) go build|' -i $(@D)/Makefile
endef

CRI_DOCKERD_AARCH64_POST_EXTRACT_HOOKS += CRI_DOCKERD_AARCH64_POST_EXTRACT_WORKAROUNDS

define CRI_DOCKERD_AARCH64_BUILD_CMDS
	$(CRI_DOCKERD_AARCH64_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) VERSION=$(CRI_DOCKERD_AARCH64_VERSION) REVISION=$(CRI_DOCKERD_AARCH64_COMMIT) cri-dockerd
endef

define CRI_DOCKERD_AARCH64_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
		$(@D)/cri-dockerd \
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
