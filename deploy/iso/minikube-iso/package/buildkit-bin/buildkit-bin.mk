################################################################################
#
# buildkit-bin
#
################################################################################

BUILDKIT_BIN_VERSION = v0.7.2
BUILDKIT_BIN_SITE = https://github.com/moby/buildkit/releases/download/$(BUILDKIT_BIN_VERSION)
BUILDKIT_BIN_SOURCE = buildkit-$(BUILDKIT_BIN_VERSION).linux-amd64.tar.gz

define BUILDKIT_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/buildctl \
		$(TARGET_DIR)/usr/bin/buildctl
	$(INSTALL) -D -m 0755 \
		$(@D)/buildkitd \
		$(TARGET_DIR)/usr/sbin/buildkitd
endef

$(eval $(generic-package))
