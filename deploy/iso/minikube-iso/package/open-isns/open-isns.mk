################################################################################
#
# open-isns
#
################################################################################

OPEN_ISNS_VERSION = v0.101
OPEN_ISNS_SITE = https://github.com/open-iscsi/open-isns/archive/refs/tags
OPEN_ISNS_SOURCE = $(OPEN_ISNS_VERSION).tar.gz

OPEN_ISNS_CFLAGS = "-I$(@D)/include -I$(@D) -D_GNU_SOURCE"

define OPEN_ISNS_BUILD_CMDS
	$(MAKE) $(TARGET_CONFIGURE_OPTS) CFLAGS=$(OPEN_ISNS_CFLAGS) -C $(@D) 
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) install DESTDIR=$(TARGET_DIR)
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) install_hdrs DESTDIR=$(TARGET_DIR)
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) install_lib DESTDIR=$(TARGET_DIR)
endef

#OPEN_ISNS_TARGET_OPTS = DESTDIR=$(TARGET_DIR) install install_hdrs install_lib
$(eval $(autotools-package))
