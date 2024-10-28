################################################################################
#
# cni-plugins-latest
#
################################################################################

CNI_PLUGINS_LATEST_VERSION = v1.6.0
CNI_PLUGINS_LATEST_SITE = https://github.com/containernetworking/plugins/releases/download/$(CNI_PLUGINS_LATEST_VERSION)
CNI_PLUGINS_LATEST_SOURCE = cni-plugins-linux-amd64-$(CNI_PLUGINS_LATEST_VERSION).tgz
CNI_PLUGINS_LATEST_LICENSE = Apache-2.0
CNI_PLUGINS_LATEST_LICENSE_FILES = LICENSE

define CNI_PLUGINS_LATEST_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/bandwidth \
		$(TARGET_DIR)/opt/cni/bin/bandwidth

	ln -sf \
		../../opt/cni/bin/bandwidth \
		$(TARGET_DIR)/usr/bin/bandwidth

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
		$(@D)/dhcp \
		$(TARGET_DIR)/opt/cni/bin/dhcp

	ln -sf \
		../../opt/cni/bin/dhcp \
		$(TARGET_DIR)/usr/bin/dhcp

	$(INSTALL) -D -m 0755 \
		$(@D)/firewall \
		$(TARGET_DIR)/opt/cni/bin/firewall

	ln -sf \
		../../opt/cni/bin/firewall \
		$(TARGET_DIR)/usr/bin/firewall
endef

$(eval $(generic-package))
