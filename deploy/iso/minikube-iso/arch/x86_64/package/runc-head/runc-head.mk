################################################################################
#
# runc
#
################################################################################

# As of 2021-12-03, v1.0.3
RUNC_HEAD_VERSION = f46b6ba2c9314cfc8caae24a32ec5fe9ef1059fe
RUNC_HEAD_SITE = https://github.com/opencontainers/runc/archive
RUNC_HEAD_SOURCE = $(RUNC_HEAD_VERSION).tar.gz
RUNC_HEAD_LICENSE = Apache-2.0
RUNC_HEAD_LICENSE_FILES = LICENSE

RUNC_HEAD_DEPENDENCIES = host-go

RUNC_HEAD_GOPATH = $(@D)/_output
RUNC_HEAD_MAKE_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=1 \
	GO111MODULE=off \
	GOPATH="$(RUNC_HEAD_GOPATH)" \
	GOBIN="$(RUNC_HEAD_GOPATH)/bin" \
	PATH=$(RUNC_HEAD_GOPATH)/bin:$(BR_PATH) \
	GOARCH=amd64

RUNC_HEAD_COMPILE_SRC = $(RUNC_HEAD_GOPATH)/src/github.com/opencontainers/runc

ifeq ($(BR2_PACKAGE_LIBSECCOMP),y)
RUNC_HEAD_GOTAGS += seccomp
RUNC_HEAD_DEPENDENCIES += libseccomp host-pkgconf
endif

define RUNC_HEAD_CONFIGURE_CMDS
	mkdir -p $(RUNC_HEAD_GOPATH)/src/github.com/opencontainers
	ln -s $(@D) $(RUNC_HEAD_GOPATH)/src/github.com/opencontainers/runc
endef

define RUNC_HEAD_BUILD_CMDS
	PWD=$(RUNC_HEAD_COMPILE_SRC) $(RUNC_HEAD_MAKE_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) BUILDTAGS="$(RUNC_HEAD_GOTAGS)" COMMIT_NO=$(RUNC_HEAD_VERSION) COMMIT=$(RUNC_HEAD_VERSION) PREFIX=/usr
endef

define RUNC_HEAD_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 $(@D)/runc $(TARGET_DIR)/usr/bin/runc
endef

$(eval $(generic-package))
