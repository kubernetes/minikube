################################################################################
#
# cni-bin
#
################################################################################

CNI_BIN_VERSION = 0.3.0
CNI_BIN_SITE = https://github.com/containernetworking/cni/releases/download/v$(CNI_BIN_VERSION)
CNI_BIN_SOURCE = cni-v$(CNI_BIN_VERSION).tgz

define CNI_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/bridge \
		$(TARGET_DIR)/bin/bridge
endef

$(eval $(generic-package))
