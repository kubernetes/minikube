################################################################################
#
# runc
#
################################################################################

# As of 2021-12-03, v1.0.3
RUNC_VERSION = f46b6ba2c9314cfc8caae24a32ec5fe9ef1059fe
RUNC_SITE = https://github.com/opencontainers/runc/archive
RUNC_SOURCE = $(RUNC_VERSION).tar.gz
RUNC_LICENSE = Apache-2.0
RUNC_LICENSE_FILES = LICENSE

RUNC_DEPENDENCIES = host-go

RUNC_GOPATH = $(@D)/_output
RUNC_MAKE_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=1 \
	GO111MODULE=off \
	GOPATH="$(RUNC_GOPATH)" \
	GOBIN="$(RUNC_GOPATH)/bin" \
	PATH=$(RUNC_GOPATH)/bin:$(BR_PATH) \
	GOARCH=arm64

RUNC_COMPILE_SRC = $(RUNC_GOPATH)/src/github.com/opencontainers/runc

ifeq ($(BR2_PACKAGE_LIBSECCOMP),y)
RUNC_GOTAGS += seccomp
RUNC_DEPENDENCIES += libseccomp host-pkgconf
endif

define RUNC_CONFIGURE_CMDS
	mkdir -p $(RUNC_GOPATH)/src/github.com/opencontainers
	ln -s $(@D) $(RUNC_GOPATH)/src/github.com/opencontainers/runc
endef

define RUNC_BUILD_CMDS
	PWD=$(RUNC_COMPILE_SRC) $(RUNC_MAKE_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) BUILDTAGS="$(RUNC_GOTAGS)" COMMIT_NO=$(RUNC_VERSION) COMMIT=$(RUNC_VERSION) PREFIX=/usr
endef

define RUNC_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 $(@D)/runc $(TARGET_DIR)/usr/bin/runc
endef

$(eval $(generic-package))
