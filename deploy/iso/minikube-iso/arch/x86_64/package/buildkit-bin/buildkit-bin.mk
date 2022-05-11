################################################################################
#
# buildkit-bin
#
################################################################################

BUILDKIT_BIN_VERSION = v0.9.0
BUILDKIT_BIN_COMMIT = c8bb937807d405d92be91f06ce2629e6202ac7a9
BUILDKIT_BIN_SITE = https://github.com/moby/buildkit/releases/download/$(BUILDKIT_BIN_VERSION)
BUILDKIT_BIN_SOURCE = buildkit-$(BUILDKIT_BIN_VERSION).linux-amd64.tar.gz

# https://github.com/opencontainers/runc.git
BUILDKIT_RUNC_VERSION = 12644e614e25b05da6fd08a38ffa0cfe1903fdec

define BUILDKIT_BIN_USERS
	- -1 buildkit -1 - - - - -
endef

define BUILDKIT_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/buildctl \
		$(TARGET_DIR)/usr/bin
	$(INSTALL) -D -m 0755 \
		$(@D)/buildkit-runc \
		$(TARGET_DIR)/usr/sbin
	$(INSTALL) -D -m 0755 \
		$(@D)/buildkit-qemu-* \
		$(TARGET_DIR)/usr/sbin
	$(INSTALL) -D -m 0755 \
		$(@D)/buildkitd \
		$(TARGET_DIR)/usr/sbin
	$(INSTALL) -D -m 644 \
		$(BUILDKIT_BIN_PKGDIR)/buildkit.conf \
		$(TARGET_DIR)/usr/lib/tmpfiles.d/buildkit.conf
	$(INSTALL) -D -m 644 \
		$(BUILDKIT_BIN_PKGDIR)/buildkitd.toml \
		$(TARGET_DIR)/etc/buildkit/buildkitd.toml
endef

define BUILDKIT_BIN_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(BUILDKIT_BIN_PKGDIR)/buildkit.service \
		$(TARGET_DIR)/usr/lib/systemd/system/buildkit.service
	$(INSTALL) -D -m 644 \
		$(BUILDKIT_BIN_PKGDIR)/buildkit.socket \
		$(TARGET_DIR)/usr/lib/systemd/system/buildkit.socket
	$(INSTALL) -D -m 644 \
		$(BUILDKIT_BIN_PKGDIR)/51-buildkit.preset \
		$(TARGET_DIR)/usr/lib/systemd/system-preset/51-buildkit.preset
endef

$(eval $(generic-package))
