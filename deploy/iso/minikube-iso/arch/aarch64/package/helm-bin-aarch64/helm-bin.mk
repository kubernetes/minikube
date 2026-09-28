################################################################################
#
# helm-bin
#
################################################################################

HELM_BIN_AARCH64_VERSION = v4.2.3
HELM_BIN_AARCH64_SITE = https://get.helm.sh
HELM_BIN_AARCH64_SOURCE = helm-$(HELM_BIN_AARCH64_VERSION)-linux-arm64.tar.gz
HELM_BIN_AARCH64_STRIP_COMPONENTS = 1

define HELM_BIN_AARCH64_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/helm \
		$(TARGET_DIR)/usr/bin/helm
endef

$(eval $(generic-package))
