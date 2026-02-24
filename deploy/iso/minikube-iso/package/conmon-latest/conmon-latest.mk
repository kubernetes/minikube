################################################################################
#
# conmon-latest
#
################################################################################

CONMON_LATEST_VERSION = 2.2.0
CONMON_LATEST_SITE = $(call github,containers,conmon,v$(CONMON_LATEST_VERSION))
CONMON_LATEST_SOURCE = conmon-${CONMON_LATEST_VERSION}.tar.gz
CONMON_LATEST_LICENSE = Apache-2.0
CONMON_LATEST_LICENSE_FILES = LICENSE

CONMON_LATEST_DEPENDENCIES = host-pkgconf libglib2

ifeq ($(BR2_PACKAGE_LIBSECCOMP)$(BR2_TOOLCHAIN_HEADERS_AT_LEAST_5_0):$(BR2_STATIC_LIBS),yy:)
CONMON_LATEST_DISABLE_SECCOMP = 0
CONMON_LATEST_DEPENDENCIES += libseccomp
else
CONMON_LATEST_DISABLE_SECCOMP = 1
endif

define CONMON_LATEST_CONFIGURE_CMDS
	printf '#!/bin/bash\necho "$(CONMON_LATEST_DISABLE_SECCOMP)"\n' > \
		$(@D)/hack/seccomp-notify.sh
	chmod +x $(@D)/hack/seccomp-notify.sh
endef

define CONMON_LATEST_BUILD_CMDS
	$(TARGET_MAKE_ENV) $(MAKE) CC="$(TARGET_CC)" \
		CFLAGS="$(TARGET_CFLAGS) -std=c99" \
		LDFLAGS="$(TARGET_LDFLAGS)" -C $(@D) bin/conmon
endef

define CONMON_LATEST_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 755 $(@D)/bin/conmon $(TARGET_DIR)/usr/bin/conmon
endef

$(eval $(generic-package))
