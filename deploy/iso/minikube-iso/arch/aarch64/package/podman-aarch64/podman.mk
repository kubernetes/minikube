PODMAN_AARCH64_VERSION = v3.4.2
PODMAN_AARCH64_COMMIT = 2ad1fd3555de12de34e20898cc2ef901f08fe5ed
PODMAN_AARCH64_SITE = https://github.com/containers/podman/archive
PODMAN_AARCH64_SOURCE = $(PODMAN_AARCH64_VERSION).tar.gz
PODMAN_AARCH64_LICENSE = Apache-2.0
PODMAN_AARCH64_LICENSE_FILES = LICENSE

PODMAN_AARCH64_DEPENDENCIES = host-go
ifeq ($(BR2_INIT_SYSTEMD),y)
# need libsystemd for journal
PODMAN_AARCH64_DEPENDENCIES += systemd
endif

PODMAN_AARCH64_GOPATH = $(@D)/_output
PODMAN_AARCH64_BIN_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=1 \
	GOPATH="$(PODMAN_AARCH64_GOPATH)" \
	PATH=$(PODMAN_AARCH64_GOPATH)/bin:$(BR_PATH) \
	GOARCH=arm64


define PODMAN_AARCH64_USERS
	- -1 podman -1 - - - - -
endef

define PODMAN_AARCH64_MOD_VENDOR_MAKEFILE
	# "build flag -mod=vendor only valid when using modules"
	sed -e 's|-mod=vendor ||' -i $(@D)/Makefile
endef

PODMAN_AARCH64_POST_EXTRACT_HOOKS += PODMAN_AARCH64_MOD_VENDOR_MAKEFILE

define PODMAN_AARCH64_CONFIGURE_CMDS
	mkdir -p $(PODMAN_AARCH64_GOPATH) && mv $(@D)/vendor $(PODMAN_AARCH64_GOPATH)/src

	mkdir -p $(PODMAN_AARCH64_GOPATH)/src/github.com/containers
	ln -sf $(@D) $(PODMAN_AARCH64_GOPATH)/src/github.com/containers/podman

	ln -sf $(@D) $(PODMAN_AARCH64_GOPATH)/src/github.com/containers/podman/v2
endef

define PODMAN_AARCH64_BUILD_CMDS
	mkdir -p $(@D)/bin
	$(PODMAN_AARCH64_BIN_ENV) CIRRUS_TAG=$(PODMAN_AARCH64_VERSION) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) GIT_COMMIT=$(PODMAN_AARCH64_COMMIT) PREFIX=/usr podman
endef

define PODMAN_AARCH64_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 $(@D)/bin/podman $(TARGET_DIR)/usr/bin/podman
	$(INSTALL) -d -m 755 $(TARGET_DIR)/etc/cni/net.d/
	$(INSTALL) -m 644 $(@D)/cni/87-podman-bridge.conflist $(TARGET_DIR)/etc/cni/net.d/87-podman-bridge.conflist
endef

define PODMAN_AARCH64_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
			$(@D)/contrib/systemd/system/podman.service \
			$(TARGET_DIR)/usr/lib/systemd/system/podman.service
	$(INSTALL) -D -m 644 \
			$(@D)/contrib/systemd/system/podman.socket \
			$(TARGET_DIR)/usr/lib/systemd/system/podman.socket

	# Allow running podman-remote as a user in the group "podman"
	$(INSTALL) -D -m 644 \
			$(PODMAN_AARCH64_PKGDIR)/override.conf \
			$(TARGET_DIR)/usr/lib/systemd/system/podman.socket.d/override.conf
	$(INSTALL) -D -m 644 \
			$(PODMAN_AARCH64_PKGDIR)/podman.conf \
			$(TARGET_DIR)/usr/lib/tmpfiles.d/podman.conf
endef

$(eval $(generic-package))
