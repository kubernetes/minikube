################################################################################
#
# conmon
#
################################################################################

CONMON_VERSION = 2.0.3
CONMON_COMMIT = eb5fa88c26fde5ce1e3f8a1d2a8a9498b2d7dbe6
CONMON_SITE = https://github.com/containers/conmon/archive
CONMON_SOURCE = CONMON_VERSION$(CONMON_VERSION).tar.gz
CONMON_LICENSE = Apache-2.0
CONMON_LICENSE_FILES = LICENSE

CONMON_DEPENDENCIES = host-pkgconf

define CONMON_BUILD_CMDS
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) GIT_COMMIT=$(CONMON_COMMIT) PREFIX=/usr
endef

define CONMON_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 $(@D)/bin/conmon $(TARGET_DIR)/usr/libexec/crio/conmon
	$(INSTALL) -Dm755 $(@D)/bin/conmon $(TARGET_DIR)/usr/libexec/podman/conmon
endef

$(eval $(generic-package))
