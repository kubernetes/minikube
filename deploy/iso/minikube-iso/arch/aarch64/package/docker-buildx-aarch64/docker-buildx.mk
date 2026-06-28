################################################################################
#
# docker-buildx
#
################################################################################

DOCKER_BUILDX_AARCH64_VERSION = v0.33.0
DOCKER_BUILDX_AARCH64_SITE = https://github.com/docker/buildx/releases/download/$(DOCKER_BUILDX_AARCH64_VERSION)
DOCKER_BUILDX_AARCH64_SOURCE = buildx-$(DOCKER_BUILDX_AARCH64_VERSION).linux-arm64

define DOCKER_BUILDX_AARCH64_EXTRACT_CMDS
	cp $(BR2_DL_DIR)/docker-buildx-aarch64/$(DOCKER_BUILDX_AARCH64_SOURCE) $(@D)/docker-buildx
endef

define DOCKER_BUILDX_AARCH64_CONFIGURE_CMDS
        mkdir -p $(TARGET_DIR)/usr/libexec/docker/cli-plugins
endef

define DOCKER_BUILDX_AARCH64_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
		$(@D)/docker-buildx \
		$(TARGET_DIR)/usr/libexec/docker/cli-plugins/docker-buildx
endef

$(eval $(generic-package))
