################################################################################
#
# tbb
#
################################################################################

TBB_VERSION = 2018_U5
TBB_SITE = $(call github,01org,tbb,$(TBB_VERSION))
TBB_INSTALL_STAGING = YES
TBB_LICENSE = Apache-2.0
TBB_LICENSE_FILES = LICENSE

TBB_SO_VERSION = 2
TBB_LIBS = libtbb libtbbmalloc libtbbmalloc_proxy
TBB_BIN_PATH = $(@D)/build/linux_*

define TBB_BUILD_CMDS
	$(MAKE) $(TARGET_CONFIGURE_OPTS) arch=$(BR2_ARCH) -C $(@D)
endef

define TBB_INSTALL_LIBS
	$(foreach lib,$(TBB_LIBS),
		$(INSTALL) -D -m 0755 $(TBB_BIN_PATH)/$(lib).so.$(TBB_SO_VERSION) \
			$(1)/usr/lib/$(lib).so.$(TBB_SO_VERSION) ;
		ln -sf $(lib).so.$(TBB_SO_VERSION) $(1)/usr/lib/$(lib).so
	)
endef

define TBB_INSTALL_STAGING_CMDS
	mkdir -p $(STAGING_DIR)/usr/include/
	cp -a $(@D)/include/* $(STAGING_DIR)/usr/include/
	$(call TBB_INSTALL_LIBS,$(STAGING_DIR))
endef

define TBB_INSTALL_TARGET_CMDS
	$(call TBB_INSTALL_LIBS,$(TARGET_DIR))
endef

$(eval $(generic-package))
