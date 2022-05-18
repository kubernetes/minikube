################################################################################
#
# buildkit-bin
#
################################################################################

BUILDKIT_BIN_AARCH64_VERSION = v0.10.3
BUILDKIT_BIN_AARCH64_COMMIT = c8d25d9a103b70dc300a4fd55e7e576472284e31
BUILDKIT_BIN_AARCH64_SITE = https://github.com/moby/buildkit/releases/download/$(BUILDKIT_BIN_AARCH64_VERSION)
BUILDKIT_BIN_AARCH64_SOURCE = buildkit-$(BUILDKIT_BIN_AARCH64_VERSION).linux-arm64.tar.gz

# https://github.com/opencontainers/runc.git
BUILDKIT_RUNC_VERSION = 12644e614e25b05da6fd08a38ffa0cfe1903fdec

define BUILDKIT_BIN_AARCH64_USERS
	- -1 buildkit -1 - - - - -
endef

define BUILDKIT_BIN_AARCH64_INSTALL_TARGET_CMDS
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
		$(BUILDKIT_BIN_AARCH64_PKGDIR)/buildkit.conf \
		$(TARGET_DIR)/usr/lib/tmpfiles.d/buildkit.conf
	$(INSTALL) -D -m 644 \
		$(BUILDKIT_BIN_AARCH64_PKGDIR)/buildkitd.toml \
		$(TARGET_DIR)/etc/buildkit/buildkitd.toml
endef

define BUILDKIT_BIN_AARCH64_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(BUILDKIT_BIN_AARCH64_PKGDIR)/buildkit.service \
		$(TARGET_DIR)/usr/lib/systemd/system/buildkit.service
	$(INSTALL) -D -m 644 \
		$(BUILDKIT_BIN_AARCH64_PKGDIR)/buildkit.socket \
		$(TARGET_DIR)/usr/lib/systemd/system/buildkit.socket
	$(INSTALL) -D -m 644 \
		$(BUILDKIT_BIN_AARCH64_PKGDIR)/51-buildkit.preset \
		$(TARGET_DIR)/usr/lib/systemd/system-preset/51-buildkit.preset
endef

$(eval $(generic-package))
