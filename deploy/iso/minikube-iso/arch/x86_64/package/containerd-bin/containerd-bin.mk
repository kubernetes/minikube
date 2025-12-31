################################################################################
#
# containerd-bin
#
################################################################################

CONTAINERD_BIN_VERSION = 2.2.1
CONTAINERD_BIN_SITE = https://github.com/containerd/containerd/releases/download/v$(CONTAINERD_BIN_VERSION)
CONTAINERD_BIN_SOURCE = containerd-$(CONTAINERD_BIN_VERSION)-linux-amd64.tar.gz
CONTAINERD_BIN_STRIP_COMPONENTS = 0

define CONTAINERD_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
		$(@D)/bin/containerd \
		$(TARGET_DIR)/usr/bin
	$(INSTALL) -Dm755 \
		$(@D)/bin/containerd-shim-runc-v2 \
		$(TARGET_DIR)/usr/bin
	$(INSTALL) -Dm755 \
		$(@D)/bin/ctr \
		$(TARGET_DIR)/usr/bin
	$(INSTALL) -Dm755 \
		$(@D)/bin/containerd-stress \
		$(TARGET_DIR)/usr/bin
	$(INSTALL) -Dm644 \
		$(CONTAINERD_BIN_PKGDIR)/config.toml \
		$(TARGET_DIR)/etc/containerd/config.toml
	$(INSTALL) -Dm644 \
		$(CONTAINERD_BIN_PKGDIR)/containerd_docker_io_hosts.toml \
		$(TARGET_DIR)/etc/containerd/certs.d/docker.io/hosts.toml
endef

define CONTAINERD_BIN_INSTALL_INIT_SYSTEMD
	$(INSTALL) -Dm644 \
		$(CONTAINERD_BIN_PKGDIR)/containerd.service \
		$(TARGET_DIR)/usr/lib/systemd/system/containerd.service
	$(INSTALL) -Dm644 \
		$(CONTAINERD_BIN_PKGDIR)/50-minikube.preset \
		$(TARGET_DIR)/usr/lib/systemd/system-preset/50-minikube.preset
endef

$(eval $(generic-package))
