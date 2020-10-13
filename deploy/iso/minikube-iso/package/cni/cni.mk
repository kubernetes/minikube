################################################################################
#
# cni
#
################################################################################

CNI_VERSION = v0.7.1
CNI_SITE = https://github.com/containernetworking/cni/archive
CNI_SOURCE = $(CNI_VERSION).tar.gz
CNI_LICENSE = Apache-2.0
CNI_LICENSE_FILES = LICENSE

CNI_DEPENDENCIES = host-go

CNI_GOPATH = $(@D)/_output
CNI_MAKE_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=0 \
	GO111MODULE=off \
	GOPATH="$(CNI_GOPATH)" \
	GOBIN="$(CNI_GOPATH)/bin" \
	PATH=$(CNI_GOPATH)/bin:$(BR_PATH)

CNI_BUILDFLAGS = -a --ldflags '-extldflags \"-static\"'

define CNI_CONFIGURE_CMDS
        mkdir -p $(CNI_GOPATH)/src/github.com/containernetworking
        ln -sf $(@D) $(CNI_GOPATH)/src/github.com/containernetworking/cni
endef

define CNI_BUILD_CMDS
	(cd $(@D); $(CNI_MAKE_ENV) go build -o bin/cnitool $(CNI_BUILDFLAGS) ./cnitool)
endef

define CNI_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/bin/cnitool \
		$(TARGET_DIR)/opt/cni/bin/cnitool

	ln -sf \
		../../opt/cni/bin/cnitool \
		$(TARGET_DIR)/usr/bin/cnitool
endef

$(eval $(generic-package))
