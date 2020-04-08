########################################################################
#
# Falco probe (driver) kernel module
#
########################################################################

FALCO_PROBE_VERSION = 0.21.0
FALCO_PROBE_SITE = https://github.com/falcosecurity/falco/archive
FALCO_PROBE_SOURCE = $(FALCO_PROBE_VERSION).tar.gz
FALCO_PROBE_DEPENDENCIES += ncurses libyaml
FALCO_PROBE_LICENSE = Apache-2.0
FALCO_PROBE_LICENSE_FILES = COPYING

# see cmake/modules/sysdig-repo/CMakeLists.txt
FALCO_PROBE_SYSDIG_VERSION = be1ea2d9482d0e6e2cb14a0fd7e08cbecf517f94
FALCO_PROBE_EXTRA_DOWNLOADS = https://github.com/draios/sysdig/archive/${FALCO_PROBE_SYSDIG_VERSION}.tar.gz

define FALCO_PROBE_SYSDIG_SRC
	sed -e 's|URL ".*"|URL "'$(FALCO_PROBE_DL_DIR)/$(FALCO_PROBE_SYSDIG_VERSION).tar.gz'"|' -i $(@D)/cmake/modules/sysdig-repo/CMakeLists.txt
endef

FALCO_PROBE_POST_EXTRACT_HOOKS += FALCO_PROBE_SYSDIG_SRC

FALCO_PROBE_CONF_OPTS = -DFALCO_VERSION=$(FALCO_PROBE_VERSION)
FALCO_PROBE_CONF_OPTS += -DUSE_BUNDLED_DEPS=ON

FALCO_PROBE_MAKE_OPTS = driver KERNELDIR=$(LINUX_DIR)
FALCO_PROBE_INSTALL_OPTS = install_driver
FALCO_PROBE_INSTALL_STAGING_OPTS = INSTALL_MOD_PATH=$(STAGING_DIR) install_driver
FALCO_PROBE_INSTALL_TARGET_OPTS = INSTALL_MOD_PATH=$(TARGET_DIR) install_driver

$(eval $(kernel-module))
$(eval $(cmake-package))
