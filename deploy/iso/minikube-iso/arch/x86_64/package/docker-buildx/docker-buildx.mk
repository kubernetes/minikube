################################################################################
# 
# docker-buildx
#   
################################################################################

DOCKER_BUILDX_VERSION = v0.33.0
DOCKER_BUILDX_SITE = https://github.com/docker/buildx/releases/download/$(DOCKER_BUILDX_VERSION)
DOCKER_BUILDX_SOURCE = buildx-$(DOCKER_BUILDX_VERSION).linux-amd64

define DOCKER_BUILDX_EXTRACT_CMDS
	cp $(BR2_DL_DIR)/docker-buildx/$(DOCKER_BUILDX_SOURCE) $(@D)/docker-buildx
endef

define DOCKER_BUILDX_CONFIGURE_CMDS
        mkdir -p $(TARGET_DIR)/usr/libexec/docker/cli-plugins
endef

define DOCKER_BUILDX_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
                $(@D)/docker-buildx \
                $(TARGET_DIR)/usr/libexec/docker/cli-plugins/docker-buildx
endef

$(eval $(generic-package))
