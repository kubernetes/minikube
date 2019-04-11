################################################################################
#
# docker-bin
#
################################################################################

DOCKER_BIN_VERSION = 18.06.3-ce
DOCKER_BIN_SITE = https://download.docker.com/linux/static/stable/x86_64
DOCKER_BIN_SOURCE = docker-$(DOCKER_BIN_VERSION).tgz

define DOCKER_BIN_USERS
	- -1 docker -1 - - - - -
endef

define DOCKER_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/docker \
		$(TARGET_DIR)/bin/docker

	$(INSTALL) -D -m 0755 \
		$(@D)/docker-containerd-shim \
		$(TARGET_DIR)/bin/docker-containerd-shim

	# TODO(tstromberg): Remove once we can upgrade to docker 18.09, which can call standard containerd.
	$(INSTALL) -D -m 0755 \
		$(@D)/docker-containerd \
		$(TARGET_DIR)/bin/docker-containerd

	# As of 2019-01, we use upstream runc so that we may update it independently of docker.
	# TODO(tstromberg): Remove once we can upgrade to docker 18.09, which can call standard runc.
	$(INSTALL) -D -m 0755 \
		$(@D)/docker-runc \
		$(TARGET_DIR)/bin/docker-runc.orig

	$(INSTALL) -D -m 0755 \
		$(@D)/docker-containerd-ctr \
		$(TARGET_DIR)/bin/docker-containerd-ctr

	$(INSTALL) -D -m 0755 \
		$(@D)/dockerd \
		$(TARGET_DIR)/bin/dockerd

	$(INSTALL) -D -m 0755 \
		$(@D)/docker-proxy \
		$(TARGET_DIR)/bin/docker-proxy
endef

define DOCKER_BIN_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL)/package/docker-bin/docker.socket \
		$(TARGET_DIR)/usr/lib/systemd/system/docker.socket

	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/docker-bin/forward.conf \
		$(TARGET_DIR)/etc/sysctl.d/forward.conf
endef

$(eval $(generic-package))
