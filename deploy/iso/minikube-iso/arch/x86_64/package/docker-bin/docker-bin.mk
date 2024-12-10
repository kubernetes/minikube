################################################################################
#
# docker-bin
#
################################################################################

DOCKER_BIN_VERSION = 27.4.0
DOCKER_BIN_SITE = https://download.docker.com/linux/static/stable/x86_64
DOCKER_BIN_SOURCE = docker-$(DOCKER_BIN_VERSION).tgz

define DOCKER_BIN_USERS
	- -1 docker -1 - - - - -
endef

define DOCKER_BIN_INSTALL_TARGET_CMDS
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
		$(DOCKER_BIN_PKGDIR)/daemon.json \
		$(TARGET_DIR)/etc/docker/daemon.json
endef

define DOCKER_BIN_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(DOCKER_BIN_PKGDIR)/docker.socket \
		$(TARGET_DIR)/usr/lib/systemd/system/docker.socket

	$(INSTALL) -D -m 644 \
		$(DOCKER_BIN_PKGDIR)/forward.conf \
		$(TARGET_DIR)/etc/sysctl.d/forward.conf
endef

$(eval $(generic-package))
