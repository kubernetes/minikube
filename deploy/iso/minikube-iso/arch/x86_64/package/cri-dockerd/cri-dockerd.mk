################################################################################
#
# cri-dockerd
#
################################################################################

CRI_DOCKERD_VER = 0.3.15
CRI_DOCKERD_REV = c1c566e
CRI_DOCKERD_VERSION = c1c566e0cc84abe6972f0bf857ecd8fe306258d9
CRI_DOCKERD_SITE = https://github.com/Mirantis/cri-dockerd/archive
CRI_DOCKERD_SOURCE = $(CRI_DOCKERD_VERSION).tar.gz

CRI_DOCKERD_DEPENDENCIES = host-go
CRI_DOCKERD_GOPATH = $(@D)/_output
CRI_DOCKERD_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=0 \
	GO111MODULE=on \
	GOPATH="$(CRI_DOCKERD_GOPATH)" \
	PATH=$(CRI_DOCKERD_GOPATH)/bin:$(BR_PATH) \
	GOARCH=amd64

CRI_DOCKERD_COMPILE_SRC = $(CRI_DOCKERD_GOPATH)/src/github.com/Mirantis/cri-dockerd
CRI_DOCKERD_BUILDFLAGS = "-ldflags '-X github.com/Mirantis/cri-dockerd/version.Version=$(CRI_DOCKERD_VER) -X github.com/Mirantis/cri-dockerd/version.GitCommit=$(CRI_DOCKERD_REV)'"

define CRI_DOCKERD_POST_EXTRACT_WORKAROUNDS
	# Set -buildvcs=false to disable VCS stamping (fails in buildroot)
	sed -i 's|go build |go build -buildvcs=false |' -i $(@D)/packaging/static/Makefile
	# Use the GOARCH environment variable that we set
	sed -i 's|GOARCH=$(ARCH) go build|GOARCH=$(GOARCH) go build|' -i $(@D)/packaging/static/Makefile
endef

CRI_DOCKERD_POST_EXTRACT_HOOKS += CRI_DOCKERD_POST_EXTRACT_WORKAROUNDS

define CRI_DOCKERD_BUILD_CMDS
	$(CRI_DOCKERD_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) LDFLAGS=$(CRI_DOCKERD_BUILDFLAGS) GO_VERSION=$(GO_VERSION) -C $(@D) VERSION=$(CRI_DOCKERD_VER) REVISION=$(CRI_DOCKERD_REV) static-linux
endef

define CRI_DOCKERD_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
		$(@D)/packaging/static/build/linux/cri-dockerd/cri-dockerd \
		$(TARGET_DIR)/usr/bin/cri-dockerd
endef

define CRI_DOCKERD_INSTALL_INIT_SYSTEMD
	$(INSTALL) -Dm644 \
		$(@D)/packaging/systemd/cri-docker.service \
		$(TARGET_DIR)/usr/lib/systemd/system/cri-docker.service
	$(INSTALL) -Dm644 \
		$(@D)/packaging/systemd/cri-docker.socket \
		$(TARGET_DIR)/usr/lib/systemd/system/cri-docker.socket
endef

$(eval $(generic-package))
