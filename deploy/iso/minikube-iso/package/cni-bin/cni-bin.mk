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
		$(TARGET_DIR)/opt/cni/bin/bridge

	ln -sf \
		../../opt/cni/bin/bridge \
		$(TARGET_DIR)/usr/bin/bridge

	$(INSTALL) -D -m 0755 \
		$(@D)/host-local \
		$(TARGET_DIR)/opt/cni/bin/host-local

	ln -sf \
		../../opt/cni/bin/host-local \
		$(TARGET_DIR)/usr/bin/host-local

	$(INSTALL) -D -m 0755 \
		$(@D)/loopback \
		$(TARGET_DIR)/opt/cni/bin/loopback

	ln -sf \
		../../opt/cni/bin/loopback \
		$(TARGET_DIR)/usr/bin/loopback
endef

$(eval $(generic-package))
