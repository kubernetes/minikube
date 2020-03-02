################################################################################
#
# cni
#
################################################################################

CNI_VERSION = v0.6.0
CNI_SITE = https://github.com/containernetworking/cni/archive
CNI_SOURCE = $(CNI_VERSION).tar.gz
CNI_LICENSE = Apache-2.0
CNI_LICENSE_FILES = LICENSE

CNI_DEPENDENCIES = host-go

CNI_MAKE_ENV = \
	CGO_ENABLED=0 \
	GO111MODULE=off

CNI_BUILDFLAGS = -a --ldflags '-extldflags \"-static\"'

define CNI_BUILD_CMDS
	(cd $(@D); $(CNI_MAKE_ENV) ./build.sh $(CNI_BUILDFLAGS))
endef

define CNI_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/bin/noop \
		$(TARGET_DIR)/opt/cni/bin/noop

	ln -sf \
		../../opt/cni/bin/noop \
		$(TARGET_DIR)/usr/bin/noop

	$(INSTALL) -D -m 0755 \
		$(@D)/bin/cnitool \
		$(TARGET_DIR)/opt/cni/bin/cnitool

	ln -sf \
		../../opt/cni/bin/cnitool \
		$(TARGET_DIR)/usr/bin/cnitool
endef

$(eval $(generic-package))
