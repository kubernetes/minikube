################################################################################
#
# nerdctl-bin
#
################################################################################

NERDCTL_BIN_AARCH64_VERSION = 2.4.2
NERDCTL_BIN_AARCH64_COMMIT = 94bc03df5247ece1781c4b115a44ef3d4b1e0e65
NERDCTL_BIN_AARCH64_SITE = https://github.com/containerd/nerdctl/releases/download/v$(NERDCTL_BIN_AARCH64_VERSION)
NERDCTL_BIN_AARCH64_SOURCE = nerdctl-$(NERDCTL_BIN_AARCH64_VERSION)-linux-arm64.tar.gz
NERDCTL_BIN_AARCH64_STRIP_COMPONENTS = 0

define NERDCTL_BIN_AARCH64_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/nerdctl \
		$(TARGET_DIR)/usr/bin/nerdctl
endef

$(eval $(generic-package))
