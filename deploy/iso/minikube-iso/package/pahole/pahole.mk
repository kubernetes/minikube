########################################################################
#
# pahole
#
########################################################################

PAHOLE_VERSION = v1.23
PAHOLE_SITE = git://git.kernel.org/pub/scm/devel/pahole/pahole.git
PAHOLE_SITE_METHOD = git
# This guy saved me:
# https://stackoverflow.com/a/50526817
# Indeed, pahole contains git submodule and relies on them to be built.
# The problem is that buildroot default behavior is to remove .git from archive.
# Thus, it is not possible to use git submodule...
PAHOLE_GIT_SUBMODULES = YES
# We want to have static pahole binary to avoid problem while using it during
# Linux kernel build.
HOST_PAHOLE_CONF_OPTS = -DBUILD_SHARED_LIBS=OFF -D__LIB=lib
PAHOLE_LICENSE = GPL-2.0
PAHOLE_LICENSE_FILES = COPYING

$(eval $(host-cmake-package))
