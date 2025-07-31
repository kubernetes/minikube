################################################################################
#
# crun-latest
#
################################################################################

CRUN_LATEST_VERSION = 1.23
CRUN_LATEST_COMMIT = 3fc25616c5974fbdc98f852b2e80ea9219cdd4df
# need the pre-generated release tarball with the git submodules and configure
CRUN_LATEST_SITE = https://github.com/containers/crun/releases/download/$(CRUN_LATEST_VERSION)
CRUN_LATEST_LICENSE = GPL-2.0
CRUN_LATEST_LICENSE_FILES = COPYING

CRUN_LATEST_DEPENDENCIES += host-python3

CRUN_LATEST_MAKE_OPTS = crun

ifeq ($(BR2_PACKAGE_LIBCAP),y)
CRUN_LATEST_DEPENDENCIES += libcap
else
CRUN_LATEST_CONF_OPTS += --disable-caps
endif

ifeq ($(BR2_PACKAGE_LIBSECCOMP),y)
CRUN_LATEST_CONF_OPTS += --enable-seccomp
CRUN_LATEST_DEPENDENCIES += libseccomp host-pkgconf
else
CRUN_LATEST_CONF_OPTS += --disable-seccomp
endif

ifeq ($(BR2_PACKAGE_SYSTEMD),y)
CRUN_LATEST_CONF_OPTS += --enable-systemd
CRUN_LATEST_DEPENDENCIES += systemd host-pkgconf
else
CRUN_LATEST_CONF_OPTS += --disable-systemd
endif

$(eval $(autotools-package))
