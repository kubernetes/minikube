################################################################################
#
# cri-o
#
################################################################################

CRIO_BIN_VERSION = v1.35.0
# Official CRI-O Static Binaries from GitHub Releases in GCS
# See "Downloads" section in release notes: https://github.com/cri-o/cri-o/releases/tag/v1.35.0
CRIO_BIN_SITE = https://storage.googleapis.com/cri-o/artifacts
CRIO_BIN_ARCH = amd64
ifeq ($(BR2_aarch64),y)
CRIO_BIN_ARCH = arm64
endif
CRIO_BIN_SOURCE = cri-o.$(CRIO_BIN_ARCH).$(CRIO_BIN_VERSION).tar.gz

define CRIO_BIN_USERS
	- -1 crio-admin -1 - - - - -
	- -1 crio       -1 - - - - -
endef

define CRIO_BIN_INSTALL_TARGET_CMDS
	mkdir -p $(TARGET_DIR)/usr/share/containers/oci/hooks.d
	mkdir -p $(TARGET_DIR)/etc/containers/oci/hooks.d
	mkdir -p $(TARGET_DIR)/etc/crio/crio.conf.d

	$(INSTALL) -Dm755 \
		$(@D)/bin/crio \
		$(TARGET_DIR)/usr/bin/crio
	$(INSTALL) -Dm755 \
		$(@D)/bin/pinns \
		$(TARGET_DIR)/usr/bin/pinns
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_PKGDIR)/crio.conf \
		$(TARGET_DIR)/etc/crio/crio.conf
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_PKGDIR)/policy.json \
		$(TARGET_DIR)/etc/containers/policy.json
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_PKGDIR)/registries.conf \
		$(TARGET_DIR)/etc/containers/registries.conf
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_PKGDIR)/02-crio.conf \
		$(TARGET_DIR)/etc/crio/crio.conf.d/02-crio.conf

	mkdir -p $(TARGET_DIR)/etc/sysconfig
	echo 'CRIO_OPTIONS="--log-level=debug"' > $(TARGET_DIR)/etc/sysconfig/crio
endef

define CRIO_BIN_INSTALL_INIT_SYSTEMD
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_PKGDIR)/crio.service \
		$(TARGET_DIR)/usr/lib/systemd/system/crio.service
	$(call link-service,crio.service)
endef

$(eval $(generic-package))
