################################################################################
#
# docker-buildx
#
################################################################################

DOCKER_BUILDX_AARCH64_VERSION = v0.13.1
DOCKER_BUILDX_AARCH64_COMMIT = 788433953af10f2a698f5c07611dddce2e08c7a0

define DOCKER_BUILDX_AARCH64_CONFIGURE_CMDS
        mkdir -p $(TARGET_DIR)/usr/libexec/docker/cli-plugins
endef

define DOCKER_BUILDX_AARCH64_INSTALL_TARGET_CMDS
	curl -Lo $(TARGET_DIR)/usr/libexec/docker/cli-plugins/docker-buildx https://github.com/docker/buildx/releases/download/$(DOCKER_BUILDX_AARCH64_VERSION)/buildx-$(DOCKER_BUILDX_AARCH64_VERSION).linux-arm64
	chmod 0755 $(TARGET_DIR)/usr/libexec/docker/cli-plugins/docker-buildx
endef

$(eval $(generic-package))
