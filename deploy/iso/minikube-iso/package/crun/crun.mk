################################################################################
#
# crun
#
################################################################################

CRUN_VERSION = 1.2
CRUN_COMMIT = 4f6c8e0583c679bfee6a899c05ac6b916022561b
# need the pre-generated release tarball with the git submodules and configure
CRUN_SITE = https://github.com/containers/crun/releases/download/$(CRUN_VERSION)
CRUN_LICENSE = GPL-2.0
CRUN_LICENSE_FILES = COPYING

CRUN_DEPENDENCIES += host-python3

CRUN_MAKE_OPTS = crun

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
