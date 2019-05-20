PODMAN_VERSION = v1.3.1
PODMAN_COMMIT = 7210727e205c333af9a2d0ed0bb66adcf92a6369
PODMAN_SITE = https://github.com/containers/libpod/archive
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
	mkdir -p $(PODMAN_GOPATH)/src/github.com/containers
	ln -sf $(@D) $(PODMAN_GOPATH)/src/github.com/containers/libpod
	mkdir -p $(PODMAN_GOPATH)/src/github.com/varlink
	ln -sf $(@D)/vendor/github.com/varlink/go $(PODMAN_GOPATH)/src/github.com/varlink/go
endef

define PODMAN_BUILD_CMDS
	mkdir -p $(@D)/bin
	$(PODMAN_BIN_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) GIT_COMMIT=$(PODMAN_COMMIT) PREFIX=/usr podman
endef

define PODMAN_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 $(@D)/bin/podman $(TARGET_DIR)/usr/bin/podman
endef

$(eval $(generic-package))
