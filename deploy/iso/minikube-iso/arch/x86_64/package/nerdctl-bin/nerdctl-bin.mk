################################################################################
# 
# nerdctl-bin
# 
################################################################################

NERDCTL_BIN_VERSION = 2.1.1
NERDCTL_BIN_COMMIT = d16f34059f5599a58d1ea7aad4e43790f4d27118
NERDCTL_BIN_SITE = https://github.com/containerd/nerdctl/releases/download/v$(NERDCTL_BIN_VERSION)
NERDCTL_BIN_SOURCE = nerdctl-$(NERDCTL_BIN_AARCH64_VERSION)-linux-amd64.tar.gz
NERDCTL_BIN_STRIP_COMPONENTS = 0

define NERDCTL_BIN_INSTALL_TARGET_CMDS
        $(INSTALL) -D -m 0755 \
                $(@D)/nerdctl \
                $(TARGET_DIR)/usr/bin/nerdctl
endef

$(eval $(generic-package))
