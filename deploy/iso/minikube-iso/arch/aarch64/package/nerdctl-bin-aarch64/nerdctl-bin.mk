################################################################################
#
# nerdctl-bin
#
################################################################################

NERDCTL_BIN_AARCH64_VERSION = 2.3.4
NERDCTL_BIN_AARCH64_COMMIT = 0ce88b9d78b13f0caebc59c6bb01885d7df24fba
NERDCTL_BIN_AARCH64_SITE = https://github.com/containerd/nerdctl/releases/download/v$(NERDCTL_BIN_AARCH64_VERSION)
NERDCTL_BIN_AARCH64_SOURCE = nerdctl-$(NERDCTL_BIN_AARCH64_VERSION)-linux-arm64.tar.gz
NERDCTL_BIN_AARCH64_STRIP_COMPONENTS = 0

define NERDCTL_BIN_AARCH64_INSTALL_TARGET_CMDS
        $(INSTALL) -D -m 0755 \
                $(@D)/nerdctl \
                $(TARGET_DIR)/usr/bin/nerdctl
endef

$(eval $(generic-package))
