################################################################################
#
# nerdctld-bin
#
################################################################################

NERDCTLD_BIN_VERSION = 0.7.0
NERDCTLD_BIN_SITE = https://github.com/afbjorklund/nerdctld/releases/download/v$(NERDCTLD_BIN_VERSION)
NERDCTLD_BIN_SOURCE = nerdctld-$(NERDCTLD_BIN_VERSION)-linux-amd64.tar.gz
NERDCTLD_BIN_STRIP_COMPONENTS = 0

define NERDCTLD_BIN_USERS
	- -1 nerdctl -1 - - - - -
endef

define NERDCTLD_BIN_INSTALL_TARGET_CMDS
        $(INSTALL) -D -m 0755 \
                $(@D)/nerdctld \
                $(TARGET_DIR)/usr/bin/nerdctld
endef

define NERDCTLD_BIN_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
			$(NERDCTLD_BIN_PKGDIR)/nerdctl.service \
			$(TARGET_DIR)/usr/lib/systemd/system/nerdctl.service
	$(INSTALL) -D -m 644 \
			$(NERDCTLD_BIN_PKGDIR)/nerdctl.socket\
			$(TARGET_DIR)/usr/lib/systemd/system/nerdctl.socket

	# Allow running docker as a user in the group "nerdctl"
	mkdir -p $(TARGET_DIR)/usr/lib/systemd/system/nerdctl.socket.d
	printf "[Socket]\nSocketMode=0660\nSocketGroup=nerdctl\n" \
	       > $(TARGET_DIR)/usr/lib/systemd/system/nerdctl.socket.d/override.conf
	mkdir -p $(TARGET_DIR)/usr/lib/systemd/system/nerdctl.service.d
	printf "[Service]\nEnvironment=CONTAINERD_NAMESPACE=k8s.io\n" \
	       > $(TARGET_DIR)/usr/lib/systemd/system/nerdctl.service.d/override.conf
endef

$(eval $(generic-package))
