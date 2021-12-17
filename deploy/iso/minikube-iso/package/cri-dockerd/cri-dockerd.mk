################################################################################
#
# cri-dockerd
#
################################################################################

# 0.2.0-dev
CRI_DOCKERD_VERSION = 542e27dee12db61d6e96d2a83a20359474a5efa2
CRI_DOCKERD_SITE = https://github.com/Mirantis/cri-dockerd/archive
CRI_DOCKERD_SOURCE = $(CRI_DOCKERD_VERSION).tar.gz

CRI_DOCKERD_DEPENDENCIES = host-go

CRI_DOCKERD_GOPATH = $(@D)/_output
CRI_DOCKERD_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=0 \
	GO111MODULE=on \
	GOPATH="$(CRI_DOCKERD_GOPATH)" \
	GOBIN="$(CRI_DOCKERD_GOPATH)/bin" \
	PATH=$(CRI_DOCKERD_GOPATH)/bin:$(BR_PATH)

CRI_DOCKERD_COMPILE_SRC = $(CRI_DOCKERD_GOPATH)/src/github.com/Mirantis/cri-dockerd

define CRI_DOCKERD_BUILD_CMDS
	$(CRI_DOCKERD_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) static-linux
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
