################################################################################
#
# docker-buildx
#
################################################################################

DOCKER_BUILDX_AARCH64_VERSION = v0.11.2
DOCKER_BUILDX_AARCH64_COMMIT = 9872040b6626fb7d87ef7296fd5b832e8cc2ad17

define DOCKER_BUILDX_AARCH64_CONFIGURE_CMDS
        mkdir -p $(TARGET_DIR)/usr/libexec/docker/cli-plugins
endef

define DOCKER_BUILDX_AARCH64_INSTALL_TARGET_CMDS
	curl -Lo $(TARGET_DIR)/usr/libexec/docker/cli-plugins/docker-buildx https://github.com/docker/buildx/releases/download/$(DOCKER_BUILDX_AARCH64_VERSION)/buildx-$(DOCKER_BUILDX_AARCH64_VERSION).linux-arm64
	chmod 0755 $(TARGET_DIR)/usr/libexec/docker/cli-plugins/docker-buildx
endef

$(eval $(generic-package))
