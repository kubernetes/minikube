################################################################################
#
# criu
#
################################################################################

CRIU_VERSION = v3.17.1
CRIU_SITE = https://github.com/checkpoint-restore/criu/archive/refs/tags
CRIU_SOURCE = $(CRIU_VERSION).tar.gz

CRIU_DEPENDENCIES = pkgconf libcap protobuf protobuf-c libnl libnet libnftnl nftables libmnl

# Ugly hacks to fix the build flags
# CFLAGS -> USERCFLAGS and delete LDFLAGS
CRIU_CONFIGURE_OPTS = $(shell echo '$(TARGET_CONFIGURE_OPTS)' | sed 's@ CFLAGS=@ USERCFLAGS=@g' | sed 's@LDFLAGS="*"@@g')

define CRIU_BUILD_CMDS
        $(MAKE) $(CRIU_CONFIGURE_OPTS) -C $(@D) criu
endef

define CRIU_INSTALL_TARGET_CMDS
        $(INSTALL) -Dm755 $(@D)/criu/criu $(TARGET_DIR)/usr/bin/criu 
endef

$(eval $(generic-package))
