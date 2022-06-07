################################################################################
#
# cni
#
################################################################################

CNI_AARCH64_VERSION = v0.7.1
CNI_AARCH64_SITE = https://github.com/containernetworking/cni/archive
CNI_AARCH64_SOURCE = $(CNI_AARCH64_VERSION).tar.gz
CNI_AARCH64_LICENSE = Apache-2.0
CNI_AARCH64_LICENSE_FILES = LICENSE

CNI_AARCH64_DEPENDENCIES = host-go

CNI_AARCH64_GOPATH = $(@D)/_output
CNI_AARCH64_MAKE_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=0 \
	GO111MODULE=off \
	GOPATH="$(CNI_AARCH64_GOPATH)" \
	PATH=$(CNI_AARCH64_GOPATH)/bin:$(BR_PATH) \
	GOARCH=arm64

CNI_AARCH64_BUILDFLAGS = -a --ldflags '-extldflags \"-static\"'

define CNI_AARCH64_CONFIGURE_CMDS
        mkdir -p $(CNI_AARCH64_GOPATH)/src/github.com/containernetworking
        ln -sf $(@D) $(CNI_AARCH64_GOPATH)/src/github.com/containernetworking/cni
endef

define CNI_AARCH64_BUILD_CMDS
	(cd $(@D); $(CNI_AARCH64_MAKE_ENV) go build -o bin/cnitool $(CNI_AARCH64_BUILDFLAGS) ./cnitool)
endef

define CNI_AARCH64_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/bin/cnitool \
		$(TARGET_DIR)/opt/cni/bin/cnitool

	ln -sf \
		../../opt/cni/bin/cnitool \
		$(TARGET_DIR)/usr/bin/cnitool
endef

$(eval $(generic-package))
