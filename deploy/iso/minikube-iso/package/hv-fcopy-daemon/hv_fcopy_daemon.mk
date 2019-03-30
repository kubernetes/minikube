################################################################################
#
# hv-fcopy-daemon
#
################################################################################

HV_FCOPY_DAEMON_VERSION = 4.15.1
HV_FCOPY_DAEMON_SITE = https://www.kernel.org/pub/linux/kernel/v${HV_FCOPY_DAEMON_VERSION%%.*}.x
HV_FCOPY_DAEMON_SOURCE = linux-$(HV_FCOPY_DAEMON_VERSION).tar.xz

define HV_FCOPY_DAEMON_BUILD_CMDS
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D)/tools/hv/
endef

define HV_FCOPY_DAEMON_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/tools/hv/hv_fcopy_daemon \
		$(TARGET_DIR)/usr/sbin/hv_fcopy_daemon
endef

define HV_FCOPY_DAEMON_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/hv-fcopy-daemon/hv_fcopy_daemon.service \
		$(TARGET_DIR)/usr/lib/systemd/system/hv_fcopy_daemon.service

	ln -fs /usr/lib/systemd/system/hv_fcopy_daemon.service \
		$(TARGET_DIR)/etc/systemd/system/multi-user.target.wants/hv_fcopy_daemon.service
endef

$(eval $(generic-package))
