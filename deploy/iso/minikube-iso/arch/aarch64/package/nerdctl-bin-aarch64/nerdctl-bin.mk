################################################################################
#
# nerdctl-bin
#
################################################################################

NERDCTL_BIN_AARCH64_VERSION = 2.4.1
NERDCTL_BIN_AARCH64_COMMIT = 9039d5272eb8646d35322b9f4413bdf18ce7231a
NERDCTL_BIN_AARCH64_SITE = https://github.com/containerd/nerdctl/releases/download/v$(NERDCTL_BIN_AARCH64_VERSION)
NERDCTL_BIN_AARCH64_SOURCE = nerdctl-$(NERDCTL_BIN_AARCH64_VERSION)-linux-arm64.tar.gz
NERDCTL_BIN_AARCH64_STRIP_COMPONENTS = 0

define NERDCTL_BIN_AARCH64_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/nerdctl \
		$(TARGET_DIR)/usr/bin/nerdctl
endef

$(eval $(generic-package))
