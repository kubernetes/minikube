################################################################################
#
# hv-vss-daemon
#
################################################################################

HV_VSS_DAEMON_VERSION = 4.15.1
HV_VSS_DAEMON_SITE = https://www.kernel.org/pub/linux/kernel/v${HV_VSS_DAEMON_VERSION%%.*}.x
HV_VSS_DAEMON_SOURCE = linux-$(HV_VSS_DAEMON_VERSION).tar.xz

define HV_VSS_DAEMON_BUILD_CMDS
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D)/tools/hv/
endef

define HV_VSS_DAEMON_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/tools/hv/hv_vss_daemon \
		$(TARGET_DIR)/usr/sbin/hv_vss_daemon
endef

define HV_VSS_DAEMON_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/hv-vss-daemon/70-hv_vss.rules \
		$(TARGET_DIR)/etc/udev/rules.d/70-hv_vss.rules

	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/hv-vss-daemon/hv_vss_daemon.service \
		$(TARGET_DIR)/usr/lib/systemd/system/hv_vss_daemon.service

	ln -fs /usr/lib/systemd/system/hv_vss_daemon.service \
		$(TARGET_DIR)/etc/systemd/system/multi-user.target.wants/hv_vss_daemon.service
endef

$(eval $(generic-package))
