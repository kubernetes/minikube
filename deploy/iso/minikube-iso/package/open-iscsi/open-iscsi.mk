################################################################################
#
# open-iscsi
#
################################################################################

OPEN_ISCSI_VERSION = 2.1.5
OPEN_ISCSI_SITE = https://github.com/open-iscsi/open-iscsi/archive/refs/tags
OPEN_ISCSI_SOURCE = $(OPEN_ISCSI_VERSION).tar.gz
OPEN_ISCSI_DEPENDENCIES += open-isns

OPEN_ISCSI_CFLAGS = "-fPIC -I$(@D)/include -I$(@D)/usr -I$(@D)/libopeniscsiusr -I$(TARGET_DIR)/usr/include/ -D_GNU_SOURCE -L$(TARGET_DIR)/usr/lib64 -lkmod -lsystemd"

define OPEN_ISCSI_BUILD_CMDS
	$(MAKE) -C $(@D) clean
	$(MAKE) $(TARGET_CONFIGURE_OPTS) CFLAGS=$(OPEN_ISCSI_CFLAGS) -C $(@D)/libopeniscsiusr install
	$(MAKE) $(TARGET_CONFIGURE_OPTS) CFLAGS=$(OPEN_ISCSI_CFLAGS) -C $(@D)/utils/sysdeps 
	$(MAKE) $(TARGET_CONFIGURE_OPTS) CFLAGS=$(OPEN_ISCSI_CFLAGS) -C $(@D)/utils/fwparam_ibft
	$(MAKE) $(TARGET_CONFIGURE_OPTS) CFLAGS=$(OPEN_ISCSI_CFLAGS) -C $(@D)/usr
	$(MAKE) $(TARGET_CONFIGURE_OPTS) CFLAGS=$(OPEN_ISCSI_CFLAGS) -C $(@D)/utils
	autoreconf $(@D)/iscsiuio --install
	cd $(@D)/iscsiuio && ./configure	
	$(MAKE) $(TARGET_CONFIGURE_OPTS) CFLAGS=$(OPEN_ISCSI_CFLAGS) -C $(@D)/iscsiuio
	$(MAKE) -C $(@D) CFLAGS=$(OPEN_ISCSI_CFLAGS) install
endef

define OPEN_ISCSI_INSTALL_TARGET_CMDS
        $(INSTALL) -Dm755 $(@D)/usr/iscsid $(TARGET_DIR)/usr/sbin/iscsid
        $(INSTALL) -Dm755 $(@D)/usr/iscsiadm $(TARGET_DIR)/usr/sbin/iscsiadm
        $(INSTALL) -Dm755 $(@D)/usr/iscsistart $(TARGET_DIR)/usr/sbin/iscsistart
        $(INSTALL) -Dm755 $(@D)/utils/iscsi-iname $(TARGET_DIR)/usr/sbin/iscsi-iname
        $(INSTALL) -Dm755 $(@D)/utils/iscsi-gen-initiatorname $(TARGET_DIR)/usr/sbin/iscsi-gen-initiatorname
        $(INSTALL) -Dm755 $(@D)/iscsiuio/src/unix/iscsiuio  $(TARGET_DIR)/usr/sbin/iscsiuio

        $(INSTALL) -Dm755 $(@D)/libopeniscsiusr/libopeniscsiusr.so.0.2.0 $(TARGET_DIR)/usr/lib64/libopeniscsiusr.so.0.2.0
	ln -s $(TARGET_DIR)/usr/lib64/libopeniscsiusr.so.0.2.0 $(TARGET_DIR)/usr/lib64/libopeniscsiusr.so

	$(INSTALL) -d $(TARGET_DIR)/etc/iscsi
	$(INSTALL) -Dm755 $(@D)/etc/iscsid.conf $(TARGET_DIR)/etc/iscsi/iscsid.conf

	echo "InitiatorName=`$(TARGET_DIR)/usr/sbin/iscsi-iname`" > $(TARGET_DIR)/etc/iscsi/initiatorname.iscsi 

endef 

define OPEN_ISCSI_INSTALL_INIT_SYSTEMD
	$(INSTALL) -Dm644 \
		$(@D)/etc/systemd/iscsiuio.socket \
		$(TARGET_DIR)/usr/lib/systemd/system/iscsiuio.socket
	$(INSTALL) -Dm644 \
		$(@D)/etc/systemd/iscsiuio.service \
		$(TARGET_DIR)/usr/lib/systemd/system/iscsiuio.service
	$(INSTALL) -Dm644 \
		$(@D)/etc/systemd/iscsi.service \
		$(TARGET_DIR)/usr/lib/systemd/system/iscsi.service
	$(INSTALL) -Dm644 \
		$(@D)/etc/systemd/iscsid.service \
		$(TARGET_DIR)/usr/lib/systemd/system/iscsid.service
	$(INSTALL) -Dm644 \
		$(@D)/etc/systemd/iscsid.socket \
		$(TARGET_DIR)/usr/lib/systemd/system/iscsid.socket
	$(INSTALL) -Dm644 \
		$(@D)/etc/systemd/iscsi-init.service \
		$(TARGET_DIR)/usr/lib/systemd/system/iscsi-init.service
endef

$(eval $(generic-package))
