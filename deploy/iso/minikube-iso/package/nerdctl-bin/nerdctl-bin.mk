################################################################################
#
# nerdctl-bin
#
################################################################################

NERDCTL_BIN_VERSION = v0.15.0
NERDCTL_BIN_COMMIT = b72b5ca14550b2e23a42787664b6182524c5053f
NERDCTL_BIN_SITE = https://github.com/containerd/nerdctl/releases/download/$(NERDCTL_BIN_VERSION)
NERDCTL_BIN_SOURCE = nerdctl-$(subst v,,$(NERDCTL_BIN_VERSION))-linux-amd64.tar.gz
NERDCTL_BIN_STRIP_COMPONENTS = 0

define NERDCTL_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/nerdctl \
		$(TARGET_DIR)/usr/bin
endef

$(eval $(generic-package))
