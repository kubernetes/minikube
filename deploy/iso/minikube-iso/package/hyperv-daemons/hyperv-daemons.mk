################################################################################
#
# hyperv-daemons
#
################################################################################

HYPERV_DAEMONS_VERSION = 4.15.1
HYPERV_DAEMONS_SITE = https://www.kernel.org/pub/linux/kernel/v${HYPERV_DAEMONS_VERSION%%.*}.x
HYPERV_DAEMONS_SOURCE = linux-$(HYPERV_DAEMONS_VERSION).tar.xz

define HYPERV_DAEMONS_BUILD_CMDS
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D)/tools/hv/
endef

define HYPERV_DAEMONS_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/tools/hv/hv_fcopy_daemon \
		$(TARGET_DIR)/usr/sbin/hv_fcopy_daemon

	$(INSTALL) -D -m 0755 \
		$(@D)/tools/hv/hv_kvp_daemon \
		$(TARGET_DIR)/usr/sbin/hv_kvp_daemon
	$(INSTALL) -D -m 0755 \
		$(@D)/tools/hv/hv_get_dhcp_info.sh \
		$(TARGET_DIR)/usr/libexec/hypervkvpd/hv_get_dhcp_info
	$(INSTALL) -D -m 0755 \
		$(@D)/tools/hv/hv_get_dns_info.sh \
		$(TARGET_DIR)/usr/libexec/hypervkvpd/hv_get_dns_info
	$(INSTALL) -D -m 0755 \
		$(@D)/tools/hv/hv_set_ifconfig.sh \
		$(TARGET_DIR)/usr/libexec/hypervkvpd/hv_set_ifconfig

	$(INSTALL) -D -m 0755 \
		$(@D)/tools/hv/hv_vss_daemon \
		$(TARGET_DIR)/usr/sbin/hv_vss_daemon
endef

define HYPERV_DAEMONS_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/hyperv-daemons/70-hv_fcopy.rules \
		$(TARGET_DIR)/etc/udev/rules.d/70-hv_fcopy.rules
	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/hyperv-daemons/70-hv_kvp.rules \
		$(TARGET_DIR)/etc/udev/rules.d/70-hv_kvp.rules
	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/hyperv-daemons/70-hv_vss.rules \
		$(TARGET_DIR)/etc/udev/rules.d/70-hv_vss.rules

	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/hyperv-daemons/hv_fcopy_daemon.service \
		$(TARGET_DIR)/usr/lib/systemd/system/hv_fcopy_daemon.service
	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/hyperv-daemons/hv_kvp_daemon.service \
		$(TARGET_DIR)/usr/lib/systemd/system/hv_kvp_daemon.service
	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/hyperv-daemons/hv_vss_daemon.service \
		$(TARGET_DIR)/usr/lib/systemd/system/hv_vss_daemon.service

	ln -fs /usr/lib/systemd/system/hv_fcopy_daemon.service \
		$(TARGET_DIR)/etc/systemd/system/multi-user.target.wants/hv_fcopy_daemon.service
	ln -fs /usr/lib/systemd/system/hv_kvp_daemon.service \
		$(TARGET_DIR)/etc/systemd/system/multi-user.target.wants/hv_kvp_daemon.service
	ln -fs /usr/lib/systemd/system/hv_vss_daemon.service \
		$(TARGET_DIR)/etc/systemd/system/multi-user.target.wants/hv_vss_daemon.service
endef

$(eval $(generic-package))
