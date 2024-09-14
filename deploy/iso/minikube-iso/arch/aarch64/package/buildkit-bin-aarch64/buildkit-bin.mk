################################################################################
#
# buildkit-bin
#
################################################################################

BUILDKIT_BIN_AARCH64_VERSION = v0.16.0
BUILDKIT_BIN_AARCH64_COMMIT = 0865fcc9b78559e856e81dc52b3613701e7be28d
BUILDKIT_BIN_AARCH64_SITE = https://github.com/moby/buildkit/releases/download/$(BUILDKIT_BIN_AARCH64_VERSION)
BUILDKIT_BIN_AARCH64_SOURCE = buildkit-$(BUILDKIT_BIN_AARCH64_VERSION).linux-arm64.tar.gz

# https://github.com/opencontainers/runc.git
BUILDKIT_RUNC_VERSION = 5fd4c4d144137e991c4acebb2146ab1483a97925

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
