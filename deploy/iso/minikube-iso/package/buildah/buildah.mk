BUILDAH_VERSION = v1.4
BUILDAH_SITE = https://github.com/containers/buildah/archive
BUILDAH_SOURCE = $(BUILDAH_VERSION).tar.gz
BUILDAH_LICENSE = Apache-2.0
BUILDAH_LICENSE_FILES = LICENSE

BUILDAH_DEPENDENCIES = host-go

BUILDAH_GOPATH = $(@D)/_output
BUILDAH_BIN_ENV = \
	CGO_ENABLED=1 \
	GOPATH="$(BUILDAH_GOPATH)" \
	GOBIN="$(BUILDAH_GOPATH)/bin" \
	PATH=$(BUILDAH_GOPATH)/bin:$(BR_PATH)


define BUILDAH_CONFIGURE_CMDS
	mkdir -p $(BUILDAH_GOPATH)
	mv $(@D)/vendor $(BUILDAH_GOPATH)/src
	mkdir -p $(BUILDAH_GOPATH)/src/github.com/containers
	ln -sf $(@D) $(BUILDAH_GOPATH)/src/github.com/containers/buildah
endef

define BUILDAH_BUILD_CMDS
	mkdir -p $(@D)/bin
	$(BUILDAH_BIN_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) PREFIX=/usr buildah
endef

define BUILDAH_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 $(@D)/buildah $(TARGET_DIR)/usr/bin/buildah
endef

$(eval $(generic-package))
