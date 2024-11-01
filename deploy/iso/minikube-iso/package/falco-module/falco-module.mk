########################################################################
#
# Falco driver kernel module
#
########################################################################

FALCO_MODULE_VERSION = 0.32.0
FALCO_MODULE_SITE = https://github.com/falcosecurity/falco/archive
FALCO_MODULE_SOURCE = $(FALCO_MODULE_VERSION).tar.gz
FALCO_MODULE_DEPENDENCIES += libyaml
FALCO_MODULE_LICENSE = Apache-2.0
FALCO_MODULE_LICENSE_FILES = COPYING

# see cmake/modules/falcosecurity-libs.cmake
FALCO_MODULE_FALCOSECURITY_LIBS_VERSION = 39ae7d40496793cf3d3e7890c9bbdc202263836b
FALCO_MODULE_EXTRA_DOWNLOADS = https://github.com/falcosecurity/libs/archive/$(FALCO_MODULE_FALCOSECURITY_LIBS_VERSION).tar.gz

define FALCO_MODULE_FALCOSECURITY_LIBS_SRC
	sed -e 's|URL ".*"|URL "'$(FALCO_MODULE_DL_DIR)/$(FALCO_MODULE_FALCOSECURITY_LIBS_VERSION).tar.gz'"|' -i $(@D)/cmake/modules/falcosecurity-libs-repo/CMakeLists.txt
endef

FALCO_MODULE_POST_EXTRACT_HOOKS += FALCO_MODULE_FALCOSECURITY_LIBS_SRC

FALCO_MODULE_CONF_OPTS = -DFALCO_VERSION=$(FALCO_MODULE_VERSION)
FALCO_MODULE_CONF_OPTS += -DUSE_BUNDLED_DEPS=ON

define FALCO_MODULE_BUILD_CMDS
	$(LINUX_MAKE_ENV) $(MAKE) $(LINUX_MAKE_FLAGS) driver KERNELDIR=$(LINUX_DIR) -C $(@D)
endef

define FALCO_MODULE_INSTALL_TARGET_CMDS
	$(LINUX_MAKE_ENV) $(MAKE) $(LINUX_MAKE_FLAGS) driver KERNELDIR=$(LINUX_DIR) INSTALL_MOD_PATH=$(TARGET_DIR) install_driver -C $(@D)
endef

$(eval $(kernel-module))
$(eval $(cmake-package))
