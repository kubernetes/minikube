################################################################################
#
# openvmtools
#
################################################################################

OPENVMTOOLS10_VERSION = stable-10.0.7
OPENVMTOOLS10_SITE = $(call github,vmware,open-vm-tools,$(OPENVMTOOLS10_VERSION))
OPENVMTOOLS10_SUBDIR = open-vm-tools
OPENVMTOOLS10_LICENSE = LGPLv2.1
OPENVMTOOLS10_LICENSE_FILES = $(OPENVMTOOLS10_SUBDIR)/COPYING
# Autoreconf needed or config/missing will run configure again at buildtime

define OPENVMTOOLS10_RUN_AUTOCONF
    cd $(@D)/open-vm-tools; $(HOST_DIR)/usr/bin/autoreconf -i
endef

OPENVMTOOLS10_PRE_PATCH_HOOKS += OPENVMTOOLS10_RUN_AUTOCONF

OPENVMTOOLS10_CONF_OPTS = --with-dnet \
	--without-icu --without-x --without-gtk2 \
	--without-gtkmm --without-kernel-modules \
	--disable-deploypkg --without-xerces
OPENVMTOOLS10_CONF_ENV += CUSTOM_DNET_CPPFLAGS=" "
OPENVMTOOLS10_DEPENDENCIES = libglib2 libdnet

# When libfuse is available, openvmtools can build vmblock-fuse, so
# make sure that libfuse gets built first
ifeq ($(BR2_PACKAGE_LIBFUSE),y)
OPENVMTOOLS10_DEPENDENCIES += libfuse
endif

ifeq ($(BR2_PACKAGE_OPENSSL),y)
OPENVMTOOLS10_CONF_OPTS += --with-ssl
OPENVMTOOLS10_DEPENDENCIES += openssl
else
OPENVMTOOLS10_CONF_OPTS += --without-ssl
endif

ifeq ($(BR2_PACKAGE_OPENVMTOOLS10_PROCPS),y)
OPENVMTOOLS10_CONF_OPTS += --with-procps
OPENVMTOOLS10_DEPENDENCIES += procps-ng
else
OPENVMTOOLS10_CONF_OPTS += --without-procps
endif

ifeq ($(BR2_PACKAGE_OPENVMTOOLS10_PAM),y)
OPENVMTOOLS10_CONF_OPTS += --with-pam
OPENVMTOOLS10_DEPENDENCIES += linux-pam
else
OPENVMTOOLS10_CONF_OPTS += --without-pam
endif

# configure needs execution permission
define OPENVMTOOLS10_PRE_CONFIGURE_CHMOD
	chmod 0755 $(@D)/$(OPENVMTOOLS10_SUBDIR)/configure
endef

OPENVMTOOLS10_PRE_CONFIGURE_HOOKS += OPENVMTOOLS10_PRE_CONFIGURE_CHMOD

# symlink needed by lib/system/systemLinux.c (or will cry in /var/log/messages)
# defined in lib/misc/hostinfoPosix.c
# /sbin/shutdown needed for Guest OS restart/shutdown from hypervisor
define OPENVMTOOLS10_POST_INSTALL_TARGET_THINGIES
	ln -fs os-release $(TARGET_DIR)/etc/lfs-release
	if [ ! -e $(TARGET_DIR)/sbin/shutdown ]; then \
		$(INSTALL) -D -m 755 package/openvmtools/shutdown \
			$(TARGET_DIR)/sbin/shutdown; \
	fi

	mkdir -p $(TARGET_DIR)/usr/local/bin

	ln -fs ../../bin/vmhgfs-fuse \
		$(TARGET_DIR)/usr/local/bin/vmhgfs-fuse
endef

OPENVMTOOLS10_POST_INSTALL_TARGET_HOOKS += OPENVMTOOLS10_POST_INSTALL_TARGET_THINGIES

define OPENVMTOOLS10_INSTALL_INIT_SYSV
	$(INSTALL) -D -m 755 package/openvmtools/S10vmtoolsd \
		$(TARGET_DIR)/etc/init.d/S10vmtoolsd
endef

define OPENVMTOOLS10_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 package/openvmtools/vmtoolsd.service \
		$(TARGET_DIR)/usr/lib/systemd/system/vmtoolsd.service
	mkdir -p $(TARGET_DIR)/etc/systemd/system/multi-user.target.wants
	ln -fs ../../../../usr/lib/systemd/system/vmtoolsd.service \
		$(TARGET_DIR)/etc/systemd/system/multi-user.target.wants/vmtoolsd.service
endef

$(eval $(autotools-package))
