################################################################################
#
# runc
#
################################################################################

# As of 2021-12-03, v1.0.3
RUNC_MASTER_AARCH64_VERSION = f46b6ba2c9314cfc8caae24a32ec5fe9ef1059fe
RUNC_MASTER_AARCH64_SITE = https://github.com/opencontainers/runc/archive
RUNC_MASTER_AARCH64_SOURCE = $(RUNC_MASTER_AARCH64_VERSION).tar.gz
RUNC_MASTER_AARCH64_LICENSE = Apache-2.0
RUNC_MASTER_AARCH64_LICENSE_FILES = LICENSE

RUNC_MASTER_AARCH64_DEPENDENCIES = host-go

RUNC_MASTER_AARCH64_GOPATH = $(@D)/_output
RUNC_MASTER_AARCH64_MAKE_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=1 \
	GO111MODULE=off \
	GOPATH="$(RUNC_MASTER_AARCH64_GOPATH)" \
	PATH=$(RUNC_MASTER_AARCH64_GOPATH)/bin:$(BR_PATH) \
	GOARCH=arm64

RUNC_MASTER_AARCH64_COMPILE_SRC = $(RUNC_MASTER_AARCH64_GOPATH)/src/github.com/opencontainers/runc

ifeq ($(BR2_PACKAGE_LIBSECCOMP),y)
RUNC_MASTER_AARCH64_GOTAGS += seccomp
RUNC_MASTER_AARCH64_DEPENDENCIES += libseccomp host-pkgconf
endif

define RUNC_MASTER_AARCH64_CONFIGURE_CMDS
	mkdir -p $(RUNC_MASTER_AARCH64_GOPATH)/src/github.com/opencontainers
	ln -s $(@D) $(RUNC_MASTER_AARCH64_GOPATH)/src/github.com/opencontainers/runc
endef

define RUNC_MASTER_AARCH64_BUILD_CMDS
	PWD=$(RUNC_MASTER_AARCH64_COMPILE_SRC) $(RUNC_MASTER_AARCH64_MAKE_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) BUILDTAGS="$(RUNC_MASTER_AARCH64_GOTAGS)" COMMIT_NO=$(RUNC_MASTER_AARCH64_VERSION) COMMIT=$(RUNC_MASTER_AARCH64_VERSION) PREFIX=/usr
endef

define RUNC_MASTER_AARCH64_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 $(@D)/runc $(TARGET_DIR)/usr/bin/runc
endef

$(eval $(generic-package))
