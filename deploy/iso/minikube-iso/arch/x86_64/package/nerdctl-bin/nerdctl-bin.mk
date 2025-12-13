################################################################################
# 
# nerdctl-bin
# 
################################################################################

NERDCTL_BIN_VERSION = 2.2.0
NERDCTL_BIN_COMMIT = 4eb4cbdb6b7ae82ab864a9829d1162a20eb61f81
NERDCTL_BIN_SITE = https://github.com/containerd/nerdctl/releases/download/v$(NERDCTL_BIN_VERSION)
NERDCTL_BIN_SOURCE = nerdctl-$(NERDCTL_BIN_AARCH64_VERSION)-linux-amd64.tar.gz
NERDCTL_BIN_STRIP_COMPONENTS = 0

define NERDCTL_BIN_INSTALL_TARGET_CMDS
        $(INSTALL) -D -m 0755 \
                $(@D)/nerdctl \
                $(TARGET_DIR)/usr/bin/nerdctl
endef

$(eval $(generic-package))
