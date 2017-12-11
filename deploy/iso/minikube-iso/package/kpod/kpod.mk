KPOD_VERSION = b85d0fa4ea5b6515088a3475a56a44c0cee5bfc5
KPOD_SITE = https://github.com/projectatomic/libpod/archive
KPOD_SOURCE = $(KPOD_VERSION).tar.gz
KPOD_LICENSE = Apache-2.0
KPOD_LICENSE_FILES = LICENSE

KPOD_DEPENDENCIES = host-go

KPOD_GOPATH = $(@D)/_output
KPOD_BIN_ENV = \
	CGO_ENABLED=1 \
	GOPATH="$(KPOD_GOPATH)" \
	PATH=$(KPOD_GOPATH)/bin:$(BR_PATH)


define KPOD_CONFIGURE_CMDS
	mkdir -p $(KPOD_GOPATH)/src/github.com/projectatomic
	ln -sf $(@D) $(KPOD_GOPATH)/src/github.com/projectatomic/libpod
	$(KPOD_BIN_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) install.tools DESTDIR=$(TARGET_DIR) PREFIX=$(TARGET_DIR)/usr
endef

define KPOD_BUILD_CMDS
	mkdir -p $(@D)/bin
	$(KPOD_BIN_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) PREFIX=/usr kpod
endef

define KPOD_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 $(@D)/bin/kpod $(TARGET_DIR)/usr/bin/kpod
endef

$(eval $(generic-package))
