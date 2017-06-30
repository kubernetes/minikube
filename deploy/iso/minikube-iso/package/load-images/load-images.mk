################################################################################
#
# minikube load-images
#
################################################################################

define LOAD_IMAGES_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/load-images/load-images.service \
		$(TARGET_DIR)/usr/lib/systemd/system/load-images.service

	ln -fs /usr/lib/systemd/system/load-images.service \
		$(TARGET_DIR)/etc/systemd/system/multi-user.target.wants/load-images.service
endef

define LOAD_IMAGES_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/load-images/load-images \
		$(TARGET_DIR)/usr/sbin/load-images
endef

$(eval $(generic-package))
