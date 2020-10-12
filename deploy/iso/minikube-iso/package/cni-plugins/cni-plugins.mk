################################################################################
#
# cni-plugins
#
################################################################################

CNI_PLUGINS_VERSION = v0.8.5
CNI_PLUGINS_SITE = https://github.com/containernetworking/plugins/archive
CNI_PLUGINS_SOURCE = $(CNI_PLUGINS_VERSION).tar.gz
CNI_PLUGINS_LICENSE = Apache-2.0
CNI_PLUGINS_LICENSE_FILES = LICENSE

CNI_PLUGINS_DEPENDENCIES = host-go

CNI_PLUGINS_MAKE_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=0 \
	GO111MODULE=off

CNI_PLUGINS_BUILDFLAGS = -a -ldflags '-extldflags -static -X github.com/containernetworking/plugins/pkg/utils/buildversion.BuildVersion=$(CNI_PLUGINS_VERSION)'


define CNI_PLUGINS_BUILD_CMDS
	(cd $(@D); $(CNI_PLUGINS_MAKE_ENV) ./build_linux.sh $(CNI_PLUGINS_BUILDFLAGS))
endef

define CNI_PLUGINS_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/bin/bandwidth \
		$(TARGET_DIR)/opt/cni/bin/bandwidth

	ln -sf \
		../../opt/cni/bin/bandwidth \
		$(TARGET_DIR)/usr/bin/bandwidth

	$(INSTALL) -D -m 0755 \
		$(@D)/bin/bridge \
		$(TARGET_DIR)/opt/cni/bin/bridge

	ln -sf \
		../../opt/cni/bin/bridge \
		$(TARGET_DIR)/usr/bin/bridge
	
	$(INSTALL) -D -m 0755 \
		$(@D)/bin/vlan \
		$(TARGET_DIR)/opt/cni/bin/vlan

	ln -sf \
		../../opt/cni/bin/vlan \
		$(TARGET_DIR)/usr/bin/vlan

	$(INSTALL) -D -m 0755 \
		$(@D)/bin/tuning \
		$(TARGET_DIR)/opt/cni/bin/tuning

	ln -sf \
		../../opt/cni/bin/tuning \
		$(TARGET_DIR)/usr/bin/tuning

	$(INSTALL) -D -m 0755 \
		$(@D)/bin/ptp \
		$(TARGET_DIR)/opt/cni/bin/ptp

	ln -sf \
		../../opt/cni/bin/ptp \
		$(TARGET_DIR)/usr/bin/ptp

	$(INSTALL) -D -m 0755 \
		$(@D)/bin/portmap \
		$(TARGET_DIR)/opt/cni/bin/portmap

	ln -sf \
		../../opt/cni/bin/portmap \
		$(TARGET_DIR)/usr/bin/portmap

	$(INSTALL) -D -m 0755 \
		$(@D)/bin/macvlan \
		$(TARGET_DIR)/opt/cni/bin/macvlan

	ln -sf \
		../../opt/cni/bin/macvlan \
		$(TARGET_DIR)/usr/bin/macvlan

	$(INSTALL) -D -m 0755 \
		$(@D)/bin/loopback \
		$(TARGET_DIR)/opt/cni/bin/loopback

	ln -sf \
		../../opt/cni/bin/loopback \
		$(TARGET_DIR)/usr/bin/loopback

	$(INSTALL) -D -m 0755 \
		$(@D)/bin/ipvlan \
		$(TARGET_DIR)/opt/cni/bin/ipvlan

	ln -sf \
		../../opt/cni/bin/ipvlan \
		$(TARGET_DIR)/usr/bin/ipvlan

	$(INSTALL) -D -m 0755 \
		$(@D)/bin/host-local \
		$(TARGET_DIR)/opt/cni/bin/host-local

	ln -sf \
		../../opt/cni/bin/host-local \
		$(TARGET_DIR)/usr/bin/host-local

	$(INSTALL) -D -m 0755 \
		$(@D)/bin/flannel \
		$(TARGET_DIR)/opt/cni/bin/flannel

	ln -sf \
		../../opt/cni/bin/flannel \
		$(TARGET_DIR)/usr/bin/flannel


	$(INSTALL) -D -m 0755 \
		$(@D)/bin/dhcp \
		$(TARGET_DIR)/opt/cni/bin/dhcp

	ln -sf \
		../../opt/cni/bin/dhcp \
		$(TARGET_DIR)/usr/bin/dhcp

	$(INSTALL) -D -m 0755 \
		$(@D)/bin/firewall \
		$(TARGET_DIR)/opt/cni/bin/firewall

	ln -sf \
		../../opt/cni/bin/firewall \
		$(TARGET_DIR)/usr/bin/firewall
endef

$(eval $(generic-package))
