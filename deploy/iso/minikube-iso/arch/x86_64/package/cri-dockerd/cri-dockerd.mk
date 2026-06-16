################################################################################
#
# cri-dockerd
#
################################################################################

CRI_DOCKERD_VERSION = 0.4.3
CRI_DOCKERD_SITE = https://github.com/Mirantis/cri-dockerd/releases/download/v$(CRI_DOCKERD_VERSION)
CRI_DOCKERD_SOURCE = cri-dockerd-$(CRI_DOCKERD_VERSION).amd64.tgz
CRI_DOCKERD_STRIP_COMPONENTS = 1

define CRI_DOCKERD_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
		$(@D)/cri-dockerd \
		$(TARGET_DIR)/usr/bin/cri-dockerd
endef

define CRI_DOCKERD_INSTALL_INIT_SYSTEMD
	$(INSTALL) -Dm644 \
		$(CRI_DOCKERD_PKGDIR)/cri-docker.service \
		$(TARGET_DIR)/usr/lib/systemd/system/cri-docker.service
	$(INSTALL) -Dm644 \
		$(CRI_DOCKERD_PKGDIR)/cri-docker.socket \
		$(TARGET_DIR)/usr/lib/systemd/system/cri-docker.socket
endef

$(eval $(generic-package))
