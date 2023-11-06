################################################################################
#
# nerdctl-bin
#
################################################################################

NERDCTL_BIN_AARCH64_VERSION = 1.7.0
NERDCTL_BIN_AARCH64_COMMIT = e674fe7ba6e49f12e88cd9c6c442e7ea5232502c
NERDCTL_BIN_AARCH64_SITE = https://github.com/containerd/nerdctl/releases/download/v$(NERDCTL_BIN_AARCH64_VERSION)
NERDCTL_BIN_AARCH64_SOURCE = nerdctl-$(NERDCTL_BIN_AARCH64_VERSION)-linux-arm64.tar.gz
NERDCTL_BIN_AARCH64_STRIP_COMPONENTS = 0

define NERDCTL_BIN_AARCH64_INSTALL_TARGET_CMDS
        $(INSTALL) -D -m 0755 \
                $(@D)/nerdctl \
                $(TARGET_DIR)/usr/bin/nerdctl
endef

$(eval $(generic-package))
