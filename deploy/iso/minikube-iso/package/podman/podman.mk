PODMAN_VERSION = v2.2.0
PODMAN_COMMIT = db1d2ff111ee9b012779ff3a5279a982520ccda4
PODMAN_SITE = https://github.com/containers/podman/archive
PODMAN_SOURCE = $(PODMAN_VERSION).tar.gz
PODMAN_LICENSE = Apache-2.0
PODMAN_LICENSE_FILES = LICENSE

PODMAN_DEPENDENCIES = host-go
ifeq ($(BR2_INIT_SYSTEMD),y)
# need libsystemd for journal
PODMAN_DEPENDENCIES += systemd
endif

PODMAN_GOPATH = $(@D)/_output
PODMAN_BIN_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=1 \
	GOPATH="$(PODMAN_GOPATH)" \
	GOBIN="$(PODMAN_GOPATH)/bin" \
	PATH=$(PODMAN_GOPATH)/bin:$(BR_PATH)


define PODMAN_USERS
	- -1 podman -1 - - - - -
endef

define PODMAN_MOD_VENDOR_MAKEFILE
	# "build flag -mod=vendor only valid when using modules"
	sed -e 's|-mod=vendor ||' -i $(@D)/Makefile
endef

PODMAN_POST_EXTRACT_HOOKS += PODMAN_MOD_VENDOR_MAKEFILE

define PODMAN_CONFIGURE_CMDS
	mkdir -p $(PODMAN_GOPATH) && mv $(@D)/vendor $(PODMAN_GOPATH)/src

	mkdir -p $(PODMAN_GOPATH)/src/github.com/containers
	ln -sf $(@D) $(PODMAN_GOPATH)/src/github.com/containers/podman

	ln -sf $(@D) $(PODMAN_GOPATH)/src/github.com/containers/podman/v2
endef

define PODMAN_BUILD_CMDS
	mkdir -p $(@D)/bin
	$(PODMAN_BIN_ENV) CIRRUS_TAG=$(PODMAN_VERSION) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) GIT_COMMIT=$(PODMAN_COMMIT) PREFIX=/usr podman
endef

define PODMAN_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 $(@D)/bin/podman $(TARGET_DIR)/usr/bin/podman
	$(INSTALL) -d -m 755 $(TARGET_DIR)/etc/cni/net.d/
	$(INSTALL) -m 644 $(@D)/cni/87-podman-bridge.conflist $(TARGET_DIR)/etc/cni/net.d/87-podman-bridge.conflist
endef

define PODMAN_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
			$(@D)/contrib/systemd/system/podman.service \
			$(TARGET_DIR)/usr/lib/systemd/system/podman.service
	$(INSTALL) -D -m 644 \
			$(@D)/contrib/systemd/system/podman.socket \
			$(TARGET_DIR)/usr/lib/systemd/system/podman.socket

	# Allow running podman-remote as a user in the group "podman"
	$(INSTALL) -D -m 644 \
			$(PODMAN_PKGDIR)/override.conf \
			$(TARGET_DIR)/usr/lib/systemd/system/podman.socket.d/override.conf
	$(INSTALL) -D -m 644 \
			$(PODMAN_PKGDIR)/podman.conf \
			$(TARGET_DIR)/usr/lib/tmpfiles.d/podman.conf
endef

$(eval $(generic-package))
