################################################################################
# 
# docker-buildx
#   
################################################################################

DOCKER_BUILDX_VERSION = v0.11.0
DOCKER_BUILDX_COMMIT = 687feca9e8dcd1534ac4c026bc4db5a49de0dd6e
DOCKER_BUILDX_SITE = https://github.com/docker/buildx/archive
DOCKER_BUILDX_SOURCE = $(DOCKER_BUILDX_VERSION).tar.gz

define DOCKER_BUILDX_CONFIGURE_CMDS
        mkdir -p $(TARGET_DIR)/usr/libexec/docker/cli-plugins
endef

define DOCKER_BUILDX_BUILD_CMDS
        $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) build
endef

define DOCKER_BUILDX_INSTALL_TARGET_CMDS
        $(INSTALL) -D -m 0755 \
                $(@D)/bin/build/docker-buildx \
                $(TARGET_DIR)/usr/libexec/docker/cli-plugins/docker-buildx
endef

$(eval $(generic-package))
