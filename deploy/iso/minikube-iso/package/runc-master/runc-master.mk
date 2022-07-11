################################################################################
#
# runc
#
################################################################################

# As of 2022-05-11, v1.1.2
RUNC_MASTER_VERSION = a916309fff0f838eb94e928713dbc3c0d0ac7aa4
RUNC_MASTER_SITE = https://github.com/opencontainers/runc/archive
RUNC_MASTER_SOURCE = $(RUNC_MASTER_VERSION).tar.gz
RUNC_MASTER_LICENSE = Apache-2.0
RUNC_MASTER_LICENSE_FILES = LICENSE

RUNC_MASTER_DEPENDENCIES = host-go

RUNC_MASTER_GOARCH=amd64
ifeq ($(BR2_aarch64),y)
RUNC_MASTER_GOARCH=arm64
endif

RUNC_MASTER_GOPATH = $(@D)/_output
RUNC_MASTER_MAKE_ENV = \
        $(GO_TARGET_ENV) \
        CGO_ENABLED=1 \
        GO111MODULE=off \
        GOPATH="$(RUNC_MASTER_GOPATH)" \
        PATH=$(RUNC_MASTER_GOPATH)/bin:$(BR_PATH) \
	GOARCH=$(RUNC_MASTER_GOARCH)

RUNC_MASTER_COMPILE_SRC = $(RUNC_MASTER_GOPATH)/src/github.com/opencontainers/runc

ifeq ($(BR2_PACKAGE_LIBSECCOMP),y)
RUNC_MASTER_GOTAGS += seccomp
RUNC_MASTER_DEPENDENCIES += libseccomp host-pkgconf
endif

define RUNC_MASTER_CONFIGURE_CMDS
        mkdir -p $(RUNC_MASTER_GOPATH)/src/github.com/opencontainers
        ln -s $(@D) $(RUNC_MASTER_GOPATH)/src/github.com/opencontainers/runc
endef

define RUNC_MASTER_BUILD_CMDS
        PWD=$(RUNC_MASTER_COMPILE_SRC) $(RUNC_MASTER_MAKE_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) BUILDTAGS="$(RUNC_MASTER_GOTAGS)" COMMIT_NO=$(RUNC_MASTER_VERSION) COMMIT=$(RUNC_MASTER_VERSION) PREFIX=/usr
endef

define RUNC_MASTER_INSTALL_TARGET_CMDS
        $(INSTALL) -D -m 0755 $(@D)/runc $(TARGET_DIR)/usr/bin/runc
endef

$(eval $(generic-package))
