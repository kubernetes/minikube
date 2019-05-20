################################################################################
#
# runc
#
################################################################################

# HEAD as of 2019-03-07
RUNC_MASTER_VERSION = 2b18fe1d885ee5083ef9f0838fee39b62d653e30
RUNC_MASTER_SITE = https://github.com/opencontainers/runc/archive
RUNC_MASTER_SOURCE = $(RUNC_MASTER_VERSION).tar.gz
RUNC_MASTER_LICENSE = Apache-2.0
RUNC_MASTER_LICENSE_FILES = LICENSE

RUNC_MASTER_DEPENDENCIES = host-go

RUNC_MASTER_GOPATH = "$(@D)/Godeps/_workspace"
RUNC_MASTER_MAKE_ENV = $(HOST_GO_TARGET_ENV) \
	CGO_ENABLED=1 \
	GOBIN="$(@D)/bin" \
	GOPATH="$(RUNC_MASTER_GOPATH)" \
	PATH=$(BR_PATH)

RUNC_MASTER_GLDFLAGS = \
	-buildmode=pie -X main.gitCommit=$(RUNC_MASTER_VERSION)

ifeq ($(BR2_STATIC_LIBS),y)
RUNC_MASTER_GLDFLAGS += -extldflags '-static'
endif

RUNC_MASTER_GOTAGS = cgo static_build

ifeq ($(BR2_PACKAGE_LIBSECCOMP),y)
RUNC_MASTER_GOTAGS += seccomp
RUNC_MASTER_DEPENDENCIES += libseccomp host-pkgconf
endif

define RUNC_MASTER_CONFIGURE_CMDS
	mkdir -p $(RUNC_MASTER_GOPATH)/src/github.com/opencontainers
	ln -s $(@D) $(RUNC_MASTER_GOPATH)/src/github.com/opencontainers/runc
endef

define RUNC_MASTER_BUILD_CMDS
	cd $(@D) && $(RUNC_MASTER_MAKE_ENV) $(HOST_DIR)/usr/bin/go \
		build -v -o $(@D)/bin/runc \
		-tags "$(RUNC_MASTER_GOTAGS)" -ldflags "$(RUNC_MASTER_GLDFLAGS)" github.com/opencontainers/runc
endef

define RUNC_MASTER_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 $(@D)/bin/runc $(TARGET_DIR)/usr/bin/runc
	# Install the binary in the location where Docker expects it, so that we can keep runc releases in sync.
	ln $(@D)/bin/runc $(TARGET_DIR)/usr/bin/docker-runc
endef

$(eval $(generic-package))
