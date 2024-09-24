################################################################################
#
# docker-bin
#
################################################################################

DOCKER_BIN_AARCH64_VERSION = 27.3.1
DOCKER_BIN_AARCH64_SITE = https://download.docker.com/linux/static/stable/aarch64
DOCKER_BIN_AARCH64_SOURCE = docker-$(DOCKER_BIN_AARCH64_VERSION).tgz

define DOCKER_BIN_AARCH64_USERS
	- -1 docker -1 - - - - -
endef

define DOCKER_BIN_AARCH64_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/docker \
		$(TARGET_DIR)/bin/docker

	# As of 2019-05, we use upstream containerd so that we may update it independently of docker.

	# As of 2019-01, we use upstream runc so that we may update it independently of docker.

	# As of 2019-05, we use upstream ctr so that we may update it independently of docker.

	$(INSTALL) -D -m 0755 \
		$(@D)/dockerd \
		$(TARGET_DIR)/bin/dockerd

	$(INSTALL) -D -m 0755 \
		$(@D)/docker-init \
		$(TARGET_DIR)/bin/docker-init

	$(INSTALL) -D -m 0755 \
		$(@D)/docker-proxy \
		$(TARGET_DIR)/bin/docker-proxy

	# https://kubernetes.io/docs/setup/production-environment/container-runtimes/#docker

	$(INSTALL) -Dm644 \
		$(DOCKER_BIN_AARCH64_PKGDIR)/daemon.json \
		$(TARGET_DIR)/etc/docker/daemon.json
endef

define DOCKER_BIN_AARCH64_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(DOCKER_BIN_AARCH64_PKGDIR)/docker.socket \
		$(TARGET_DIR)/usr/lib/systemd/system/docker.socket

	$(INSTALL) -D -m 644 \
		$(DOCKER_BIN_AARCH64_PKGDIR)/forward.conf \
		$(TARGET_DIR)/etc/sysctl.d/forward.conf
endef

$(eval $(generic-package))
