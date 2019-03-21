################################################################################
#
# rktlet
#
################################################################################

# HEAD as of 2018-05-28
RKTLET_MASTER_COMMIT = fd7fc6bf4a25f03c22e5f6e30f3d9f12c468afcb
RKTLET_MASTER_VERSION = v0.1.0-21-gfd7fc6b
RKTLET_MASTER_SITE = https://github.com/kubernetes-incubator/rktlet/archive
RKTLET_MASTER_SOURCE = $(RKTLET_MASTER_COMMIT).tar.gz
RKTLET_MASTER_LICENSE = Apache-2.0
RKTLET_MASTER_LICENSE_FILES = LICENSE

RKTLET_MASTER_DEPENDENCIES = host-go

RKTLET_MASTER_GOPATH = "$(@D)/Godeps/_workspace"
RKTLET_MASTER_MAKE_ENV = $(HOST_GO_TARGET_ENV) \
	CGO_ENABLED=1 \
	GOBIN="$(@D)/bin" \
	GOPATH="$(RKTLET_MASTER_GOPATH)" \
	PATH=$(BR_PATH)

RKTLET_MASTER_GLDFLAGS = \
	-buildmode=pie -X github.com/kubernetes-incubator/rktlet/version.Version=$(RKTLET_MASTER_VERSION)

ifeq ($(BR2_STATIC_LIBS),y)
RKTLET_MASTER_GLDFLAGS += -extldflags '-static'
endif

RKTLET_MASTER_GOTAGS = cgo static_build

ifeq ($(BR2_PACKAGE_LIBSECCOMP),y)
RKTLET_MASTER_GOTAGS += seccomp
RKTLET_MASTER_DEPENDENCIES += libseccomp host-pkgconf
endif

define RKTLET_MASTER_CONFIGURE_CMDS
	mkdir -p $(RKTLET_MASTER_GOPATH)/src/github.com/kubernetes-incubator
	ln -s $(@D) $(RKTLET_MASTER_GOPATH)/src/github.com/kubernetes-incubator/rktlet
endef

define RKTLET_MASTER_BUILD_CMDS
	cd $(@D) && $(RKTLET_MASTER_MAKE_ENV) $(HOST_DIR)/usr/bin/go \
		build -v -o $(@D)/bin/rktlet \
		-tags "$(RKTLET_MASTER_GOTAGS)" -ldflags "$(RKTLET_MASTER_GLDFLAGS)" github.com/kubernetes-incubator/rktlet/cmd/server
endef

define RKTLET_MASTER_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 $(@D)/bin/rktlet $(TARGET_DIR)/usr/bin/rktlet
endef

define RKTLET_MASTER_INSTALL_INIT_SYSTEMD
        $(INSTALL) -Dm644 \
                $(BR2_EXTERNAL_MINIKUBE_PATH)/package/rktlet-master/rktlet.service \
                $(TARGET_DIR)/usr/lib/systemd/system/rktlet.service
endef

$(eval $(generic-package))
