################################################################################
#
# minikube automount
#
################################################################################

define AUTOMOUNT_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL)/package/automount/minikube-automount.service \
		$(TARGET_DIR)/usr/lib/systemd/system/minikube-automount.service

	ln -fs /usr/lib/systemd/system/minikube-automount.service \
		$(TARGET_DIR)/etc/systemd/system/multi-user.target.wants/minikube-automount.service
endef

define AUTOMOUNT_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
		$(BR2_EXTERNAL)/package/automount/minikube-automount \
		$(TARGET_DIR)/usr/sbin/minikube-automount
endef

$(eval $(generic-package))
