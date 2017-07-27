################################################################################
#
# cni-plugins-bin
#
################################################################################

CNI_PLUGINS_BIN_VERSION = v0.6.0-rc1
CNI_PLUGINS_BIN_SITE = https://github.com/containernetworking/plugins/releases/download/$(CNI_PLUGINS_BIN_VERSION)
CNI_PLUGINS_BIN_SOURCE = cni-plugins-amd64-$(CNI_PLUGINS_BIN_VERSION).tgz

define CNI_PLUGINS_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/bridge \
		$(TARGET_DIR)/opt/cni/bin/bridge

	ln -sf \
		../../opt/cni/bin/bridge \
		$(TARGET_DIR)/usr/bin/bridge
	
	$(INSTALL) -D -m 0755 \
		$(@D)/vlan \
		$(TARGET_DIR)/opt/cni/bin/vlan

	ln -sf \
		../../opt/cni/bin/vlan \
		$(TARGET_DIR)/usr/bin/vlan

	$(INSTALL) -D -m 0755 \
		$(@D)/tuning \
		$(TARGET_DIR)/opt/cni/bin/tuning

	ln -sf \
		../../opt/cni/bin/tuning \
		$(TARGET_DIR)/usr/bin/tuning

	$(INSTALL) -D -m 0755 \
		$(@D)/sample \
		$(TARGET_DIR)/opt/cni/bin/sample

	ln -sf \
		../../opt/cni/bin/sample \
		$(TARGET_DIR)/usr/bin/sample

	$(INSTALL) -D -m 0755 \
		$(@D)/ptp \
		$(TARGET_DIR)/opt/cni/bin/ptp

	ln -sf \
		../../opt/cni/bin/ptp \
		$(TARGET_DIR)/usr/bin/ptp

	$(INSTALL) -D -m 0755 \
		$(@D)/portmap \
		$(TARGET_DIR)/opt/cni/bin/portmap

	ln -sf \
		../../opt/cni/bin/portmap \
		$(TARGET_DIR)/usr/bin/portmap

	$(INSTALL) -D -m 0755 \
		$(@D)/macvlan \
		$(TARGET_DIR)/opt/cni/bin/macvlan

	ln -sf \
		../../opt/cni/bin/macvlan \
		$(TARGET_DIR)/usr/bin/macvlan

	$(INSTALL) -D -m 0755 \
		$(@D)/loopback \
		$(TARGET_DIR)/opt/cni/bin/loopback

	ln -sf \
		../../opt/cni/bin/loopback \
		$(TARGET_DIR)/usr/bin/loopback

	$(INSTALL) -D -m 0755 \
		$(@D)/ipvlan \
		$(TARGET_DIR)/opt/cni/bin/ipvlan

	ln -sf \
		../../opt/cni/bin/ipvlan \
		$(TARGET_DIR)/usr/bin/ipvlan

	$(INSTALL) -D -m 0755 \
		$(@D)/host-local \
		$(TARGET_DIR)/opt/cni/bin/host-local

	ln -sf \
		../../opt/cni/bin/host-local \
		$(TARGET_DIR)/usr/bin/host-local

	$(INSTALL) -D -m 0755 \
		$(@D)/flannel \
		$(TARGET_DIR)/opt/cni/bin/flannel

	ln -sf \
		../../opt/cni/bin/flannel \
		$(TARGET_DIR)/usr/bin/flannel


	$(INSTALL) -D -m 0755 \
		$(@D)/dhcp \
		$(TARGET_DIR)/opt/cni/bin/dhcp

	ln -sf \
		../../opt/cni/bin/dhcp \
		$(TARGET_DIR)/usr/bin/dhcp
endef

$(eval $(generic-package))
