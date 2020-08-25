########################################################################
#
# Falco driver kernel module
#
########################################################################

FALCO_MODULE_VERSION = 0.24.0
FALCO_MODULE_SITE = https://github.com/falcosecurity/falco/archive
FALCO_MODULE_SOURCE = $(FALCO_MODULE_VERSION).tar.gz
FALCO_MODULE_DEPENDENCIES += ncurses libyaml
FALCO_MODULE_LICENSE = Apache-2.0
FALCO_MODULE_LICENSE_FILES = COPYING

# see cmake/modules/sysdig-repo/CMakeLists.txt
FALCO_MODULE_SYSDIG_VERSION = 85c88952b018fdbce2464222c3303229f5bfcfad
FALCO_MODULE_EXTRA_DOWNLOADS = https://github.com/draios/sysdig/archive/${FALCO_MODULE_SYSDIG_VERSION}.tar.gz

define FALCO_MODULE_SYSDIG_SRC
	sed -e 's|URL ".*"|URL "'$(FALCO_MODULE_DL_DIR)/$(FALCO_MODULE_SYSDIG_VERSION).tar.gz'"|' -i $(@D)/cmake/modules/sysdig-repo/CMakeLists.txt
endef

FALCO_MODULE_POST_EXTRACT_HOOKS += FALCO_MODULE_SYSDIG_SRC

FALCO_MODULE_CONF_OPTS = -DFALCO_VERSION=$(FALCO_MODULE_VERSION)
FALCO_MODULE_CONF_OPTS += -DUSE_BUNDLED_DEPS=ON

FALCO_MODULE_MAKE_ENV = $(LINUX_MAKE_ENV)
FALCO_MODULE_MAKE_OPTS = $(LINUX_MAKE_FLAGS) driver KERNELDIR=$(LINUX_DIR)
FALCO_MODULE_INSTALL_OPTS = install_driver
FALCO_MODULE_INSTALL_STAGING_OPTS = INSTALL_MOD_PATH=$(STAGING_DIR) install_driver
FALCO_MODULE_INSTALL_TARGET_OPTS = INSTALL_MOD_PATH=$(TARGET_DIR) install_driver

$(eval $(kernel-module))
$(eval $(cmake-package))
