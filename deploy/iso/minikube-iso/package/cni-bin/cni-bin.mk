################################################################################
#
# cni-bin
#
################################################################################

CNI_BIN_VERSION = v0.6.0-rc1
CNI_BIN_SITE = https://github.com/containernetworking/cni/releases/download/$(CNI_BIN_VERSION)
CNI_BIN_SOURCE = cni-amd64-$(CNI_BIN_VERSION).tgz

define CNI_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/noop \
		$(TARGET_DIR)/opt/cni/bin/noop

	ln -sf \
		../../opt/cni/bin/noop \
		$(TARGET_DIR)/usr/bin/noop

	$(INSTALL) -D -m 0755 \
		$(@D)/noop \
		$(TARGET_DIR)/opt/cni/bin/cnitool

	ln -sf \
		../../opt/cni/bin/cnitool \
		$(TARGET_DIR)/usr/bin/cnitool
endef

$(eval $(generic-package))
