PODMAN_VERSION = v0.4.1
PODMAN_SITE = https://github.com/projectatomic/libpod/archive
PODMAN_SOURCE = $(PODMAN_VERSION).tar.gz
PODMAN_LICENSE = Apache-2.0
PODMAN_LICENSE_FILES = LICENSE

PODMAN_DEPENDENCIES = host-go

PODMAN_GOPATH = $(@D)/_output
PODMAN_BIN_ENV = \
	CGO_ENABLED=1 \
	GOPATH="$(PODMAN_GOPATH)" \
	GOBIN="$(PODMAN_GOPATH)/bin" \
	PATH=$(PODMAN_GOPATH)/bin:$(BR_PATH)


define PODMAN_CONFIGURE_CMDS
	mkdir -p $(PODMAN_GOPATH)/src/github.com/projectatomic
	ln -sf $(@D) $(PODMAN_GOPATH)/src/github.com/projectatomic/libpod
	$(PODMAN_BIN_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) install.tools DESTDIR=$(TARGET_DIR) PREFIX=$(TARGET_DIR)/usr
endef

define PODMAN_BUILD_CMDS
	mkdir -p $(@D)/bin
	$(PODMAN_BIN_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) PREFIX=/usr podman
endef

define PODMAN_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 $(@D)/bin/podman $(TARGET_DIR)/usr/bin/podman
endef

$(eval $(generic-package))
