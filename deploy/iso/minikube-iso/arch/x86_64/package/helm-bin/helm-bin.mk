################################################################################
#
# helm-bin
#
################################################################################

HELM_BIN_VERSION = v4.2.3
HELM_BIN_SITE = https://get.helm.sh
HELM_BIN_SOURCE = helm-$(HELM_BIN_VERSION)-linux-amd64.tar.gz
HELM_BIN_STRIP_COMPONENTS = 1

define HELM_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/helm \
		$(TARGET_DIR)/usr/bin/helm
endef

$(eval $(generic-package))
