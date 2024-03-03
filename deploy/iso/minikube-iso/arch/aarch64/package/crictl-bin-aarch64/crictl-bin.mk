################################################################################
#
# crictl-bin
#
################################################################################

CRICTL_BIN_AARCH64_VERSION = v1.28.0
CRICTL_BIN_AARCH64_SITE = https://github.com/kubernetes-sigs/cri-tools/releases/download/$(CRICTL_BIN_AARCH64_VERSION)
CRICTL_BIN_AARCH64_SOURCE = crictl-$(CRICTL_BIN_AARCH64_VERSION)-linux-arm64.tar.gz
CRICTL_BIN_AARCH64_STRIP_COMPONENTS = 0

define CRICTL_BIN_AARCH64_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/crictl \
		$(TARGET_DIR)/usr/bin/crictl
endef

$(eval $(generic-package))
