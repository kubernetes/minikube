################################################################################
#
# sysdig
#
################################################################################

SYSDIG_VERSION = 0.27.1
SYSDIG_SITE = $(call github,draios,sysdig,$(SYSDIG_VERSION))
SYSDIG_LICENSE = GPL-2.0
SYSDIG_LICENSE_FILES = COPYING
SYSDIG_CPE_ID_VENDOR = sysdig
SYSDIG_CONF_OPTS = -DENABLE_DKMS=OFF -DUSE_BUNDLED_DEPS=OFF
SYSDIG_SUPPORTS_IN_SOURCE_BUILD = NO

SYSDIG_DEPENDENCIES = \
	c-ares \
	elfutils \
	gtest \
	grpc \
	jq \
	jsoncpp \
	libb64 \
	libcurl \
	luainterpreter \
	ncurses \
	openssl \
	protobuf \
	tbb \
	zlib

# sysdig creates the module Makefile from a template, which contains a
# single place-holder, KBUILD_FLAGS, wich is only replaced with two
# things:
#   - debug flags, which we don't care about here,
#   - 'sysdig-feature' flags, which are never set, so always empty
# So, just replace the place-holder with the only meaningful value: nothing.
define SYSDIG_MODULE_GEN_MAKEFILE
	$(INSTALL) -m 0644 $(@D)/driver/Makefile.in $(@D)/driver/Makefile
	$(SED) 's/@KBUILD_FLAGS@//;' $(@D)/driver/Makefile
	$(SED) 's/@PROBE_NAME@/sysdig-probe/;' $(@D)/driver/Makefile
endef
SYSDIG_POST_PATCH_HOOKS += SYSDIG_MODULE_GEN_MAKEFILE

# Don't build the driver as part of the 'standard' procedure, we'll
# build it on our own with the kernel-module infra.
SYSDIG_CONF_OPTS += -DBUILD_DRIVER=OFF

SYSDIG_MODULE_SUBDIRS = driver
SYSDIG_MODULE_MAKE_OPTS = KERNELDIR=$(LINUX_DIR)

$(eval $(kernel-module))
$(eval $(cmake-package))
