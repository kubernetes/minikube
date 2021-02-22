################################################################################
#
# minikube scheduled-stop
#
################################################################################

define SCHEDULED_STOP_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(SCHEDULED_STOP_PKGDIR)/minikube-scheduled-stop.service \
		$(TARGET_DIR)/usr/lib/systemd/system/minikube-scheduled-stop.service
endef

define SCHEDULED_STOP_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
		$(SCHEDULED_STOP_PKGDIR)/minikube-scheduled-stop \
		$(TARGET_DIR)/usr/sbin/minikube-scheduled-stop
endef

$(eval $(generic-package))
