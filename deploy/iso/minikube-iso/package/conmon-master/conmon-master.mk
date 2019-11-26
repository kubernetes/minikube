################################################################################
#
# conmon
#
################################################################################

# HEAD as of 2019-11-11
CONMON_MASTER_VERSION = eb5fa88c26fde5ce1e3f8a1d2a8a9498b2d7dbe6
CONMON_MASTER_SITE = https://github.com/containers/conmon/archive
CONMON_MASTER_SOURCE = $(CONMON_MASTER_VERSION).tar.gz
CONMON_MASTER_LICENSE = Apache-2.0
CONMON_MASTER_LICENSE_FILES = LICENSE

CONMON_MASTER_DEPENDENCIES = host-pkgconf

define CONMON_MASTER_BUILD_CMDS
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) GIT_COMMIT=$(CONMON_MASTER_VERSION) PREFIX=/usr
endef

define CONMON_MASTER_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 $(@D)/bin/conmon $(TARGET_DIR)/usr/libexec/crio/conmon
	$(INSTALL) -Dm755 $(@D)/bin/conmon $(TARGET_DIR)/usr/libexec/podman/conmon
endef

$(eval $(generic-package))
