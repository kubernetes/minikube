################################################################################
#
# VirtualBox Mount Service
#
################################################################################
VBOX_MOUNT_SERVICE_VERSION = 5.2.42
VBOX_MOUNT_SERVICE_SITE = http://download.virtualbox.org/virtualbox/$(VBOX_MOUNT_SERVICE_VERSION)
VBOX_MOUNT_SERVICE_LICENSE = GPLv2
VBOX_MOUNT_SERVICE_LICENSE_FILES = COPYING
VBOX_MOUNT_SERVICE_SOURCE = VirtualBox-$(VBOX_MOUNT_SERVICE_VERSION).tar.bz2
VBOX_MOUNT_SERVICE_EXTRA_DOWNLOADS = http://download.virtualbox.org/virtualbox/${VBOX_MOUNT_SERVICE_VERSION}/VBoxGuestAdditions_${VBOX_MOUNT_SERVICE_VERSION}.iso

define VBOX_MOUNT_SERVICE_USERS
	- -1 vboxsf -1 - - - - -
endef

define VBOX_MOUNT_SERVICE_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(VBOX_MOUNT_SERVICE_PKGDIR)/vboxservice.service \
		$(TARGET_DIR)/usr/lib/systemd/system/vboxservice.service

	ln -fs /usr/lib/systemd/system/vboxservice.service \
		$(TARGET_DIR)/etc/systemd/system/multi-user.target.wants/vboxservice.service
endef

define VBOX_MOUNT_SERVICE_BUILD_CMDS
	7z x -aoa $(BR2_DL_DIR)/vbox-mount-service/VBoxGuestAdditions_${VBOX_MOUNT_SERVICE_VERSION}.iso -ir'!VBoxLinuxAdditions.run' -o"$(@D)"
	sh $(@D)/VBoxLinuxAdditions.run --noexec --target $(@D)
	tar --overwrite -C $(@D) -xjf $(@D)/VBoxGuestAdditions-amd64.tar.bz2 sbin/VBoxService
	tar --overwrite -C $(@D) -xjf $(@D)/VBoxGuestAdditions-amd64.tar.bz2 bin/VBoxControl
endef

define VBOX_MOUNT_SERVICE_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m755 \
		$(@D)/sbin/VBoxService \
		$(TARGET_DIR)/sbin
	$(INSTALL) -D -m755 \
		$(@D)/bin/VBoxControl \
		$(TARGET_DIR)/sbin
endef

$(eval $(generic-package))
