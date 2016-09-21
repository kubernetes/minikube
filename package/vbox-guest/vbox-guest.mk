################################################################################
#
# VirtualBox Linux Guest Drivers
#
################################################################################

VBOX_GUEST_VERSION = 5.1.6
VBOX_GUEST_SOURCE = VirtualBox-$(VBOX_GUEST_VERSION).tar.bz2
VBOX_GUEST_SITE = http://download.virtualbox.org/virtualbox/$(VBOX_GUEST_VERSION)
VBOX_GUEST_LICENSE = GPLv2
VBOX_GUEST_LICENSE_FILES = COPYING

define VBOX_GUEST_EXPORT_MODULES
	( cd $(@D)/src/VBox/Additions/linux; ./export_modules modules.tar.gz )
	mkdir -p $(@D)/vbox-modules
	tar -C $(@D)/vbox-modules -xzf $(@D)/src/VBox/Additions/linux/modules.tar.gz
endef

VBOX_GUEST_POST_EXTRACT_HOOKS += VBOX_GUEST_EXPORT_MODULES

VBOX_GUEST_MODULE_SUBDIRS = vbox-modules
VBOX_GUEST_MODULE_MAKE_OPTS = KVERSION=$(LINUX_VERSION_PROBED)

define VBOX_GUEST_BUILD_CMDS
	$(TARGET_CC) -Wall -O2 -D_GNU_SOURCE -DIN_RING3 \
		-I$(@D)/vbox-modules/vboxsf/include \
		-I$(@D)/vbox-modules/vboxsf \
		-o $(@D)/vbox-modules/mount.vboxsf \
		$(@D)/src/VBox/Additions/linux/sharedfolders/vbsfmount.c \
		$(@D)/src/VBox/Additions/linux/sharedfolders/mount.vboxsf.c
endef

define VBOX_GUEST_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 $(@D)/vbox-modules/mount.vboxsf $(TARGET_DIR)/sbin
endef

$(eval $(kernel-module))
$(eval $(generic-package))
