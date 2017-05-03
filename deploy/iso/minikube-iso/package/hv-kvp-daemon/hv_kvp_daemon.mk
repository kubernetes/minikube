################################################################################
#
# hv-kvp-daemon
#
################################################################################

HV_KVP_DAEMON_VERSION = 4.4.27
HV_KVP_DAEMON_SITE = https://www.kernel.org/pub/linux/kernel/v${HV_KVP_DAEMON_VERSION%%.*}.x
HV_KVP_DAEMON_SOURCE = linux-$(HV_KVP_DAEMON_VERSION).tar.xz

define HV_KVP_DAEMON_BUILD_CMDS
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D)/tools/hv/
endef

define HV_KVP_DAEMON_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/tools/hv/hv_kvp_daemon \
		$(TARGET_DIR)/usr/sbin/hv_kvp_daemon
endef

define HV_KVP_DAEMON_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/hv-kvp-daemon/hv_kvp_daemon.service \
		$(TARGET_DIR)/usr/lib/systemd/system/hv_kvp_daemon.service

	ln -fs /usr/lib/systemd/system/hv_kvp_daemon.service \
		$(TARGET_DIR)/etc/systemd/system/multi-user.target.wants/hv_kvp_daemon.service
endef

$(eval $(generic-package))
