################################################################################
#
# crun
#
################################################################################

CRUN_VERSION = 1.13
CRUN_COMMIT = c761349704905da07cfe67f38dda6850334a160f
# need the pre-generated release tarball with the git submodules and configure
CRUN_SITE = https://github.com/containers/crun/releases/download/$(CRUN_VERSION)
CRUN_LICENSE = GPL-2.0
CRUN_LICENSE_FILES = COPYING

CRUN_DEPENDENCIES += host-python3

CRUN_MAKE_OPTS = crun

define CRUN_GIT_VERSION_FIXUP
	cp $(CRUN_DIR)/.tarball-git-version.h $(CRUN_DIR)/git-version.h
endef
CRUN_POST_EXTRACT_HOOKS += CRUN_GIT_VERSION_FIXUP

ifeq ($(BR2_PACKAGE_LIBCAP),y)
CRUN_DEPENDENCIES += libcap
else
CRUN_CONF_OPTS += --disable-caps
endif

ifeq ($(BR2_PACKAGE_LIBSECCOMP),y)
CRUN_CONF_OPTS += --enable-seccomp
CRUN_DEPENDENCIES += libseccomp host-pkgconf
else
CRUN_CONF_OPTS += --disable-seccomp
endif

ifeq ($(BR2_PACKAGE_SYSTEMD),y)
CRUN_CONF_OPTS += --enable-systemd
CRUN_DEPENDENCIES += systemd host-pkgconf
else
CRUN_CONF_OPTS += --disable-systemd
endif

$(eval $(autotools-package))
