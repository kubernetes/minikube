################################################################################
#
# runc
#
################################################################################

RUNC_VERSION = 1.0.0-rc9
RUNC_COMMIT = d736ef14f0288d6993a1845745d6756cfc9ddd5a
RUNC_SITE = https://github.com/opencontainers/runc/archive
RUNC_SOURCE = v$(RUNC_VERSION).tar.gz
RUNC_LICENSE = Apache-2.0
RUNC_LICENSE_FILES = LICENSE

RUNC_DEPENDENCIES = host-go

RUNC_GOPATH = "$(@D)/Godeps/_workspace"
RUNC_MAKE_ENV = $(HOST_GO_TARGET_ENV) \
	CGO_ENABLED=1 \
	GOBIN="$(@D)/bin" \
	GOPATH="$(RUNC_GOPATH)" \
	PATH=$(BR_PATH)

RUNC_GLDFLAGS = \
	-buildmode=pie -X main.gitCommit=$(RUNC_COMMIT) -X main.version=$(RUNC_VERSION)

ifeq ($(BR2_PACKAGE_LIBSECCOMP),y)
RUNC_GOTAGS += seccomp
RUNC_DEPENDENCIES += libseccomp host-pkgconf
endif

define RUNC_CONFIGURE_CMDS
	mkdir -p $(RUNC_GOPATH)/src/github.com/opencontainers
	ln -s $(@D) $(RUNC_GOPATH)/src/github.com/opencontainers/runc
endef

define RUNC_BUILD_CMDS
	cd $(@D) && $(RUNC_MAKE_ENV) $(HOST_DIR)/usr/bin/go \
		build -v -o $(@D)/bin/runc \
		-tags "$(RUNC_GOTAGS)" -ldflags "$(RUNC_GLDFLAGS)" github.com/opencontainers/runc
endef

define RUNC_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 $(@D)/bin/runc $(TARGET_DIR)/usr/bin/runc
endef

$(eval $(generic-package))
