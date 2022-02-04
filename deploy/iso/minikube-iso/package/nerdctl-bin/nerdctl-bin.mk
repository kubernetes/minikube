################################################################################
#
# nerdctl-bin
#
################################################################################

NERDCTL_BIN_VERSION = v0.16.1
NERDCTL_BIN_COMMIT = c4bd56b3aa220db037cc6c0a4e0c8cc062f2cc4c
NERDCTL_BIN_SITE = https://github.com/containerd/nerdctl/releases/download/$(NERDCTL_BIN_VERSION)
NERDCTL_BIN_SOURCE = nerdctl-$(subst v,,$(NERDCTL_BIN_VERSION))-linux-amd64.tar.gz
NERDCTL_BIN_STRIP_COMPONENTS = 0

define NERDCTL_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/nerdctl \
		$(TARGET_DIR)/usr/bin
endef

$(eval $(generic-package))
