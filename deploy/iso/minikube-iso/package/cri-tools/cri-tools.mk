################################################################################
#
# cri-tools
#
################################################################################

CRI_TOOLS_VERSION = v1.11.1
CRI_TOOLS_SITE = https://github.com/kubernetes-incubator/cri-tools/archive
CRI_TOOLS_SOURCE = $(CRI_TOOLS_VERSION).tar.gz
CRI_TOOLS_LICENSE = Apache-2.0
CRI_TOOLS_LICENSE_FILES = LICENSE
CRI_TOOLS_DEPENDENCIES = host-go
CRI_TOOLS_GOPATH = $(@D)/_output
CRI_TOOLS_ENV = \
	CGO_ENABLED=1 \
	GOPATH="$(CRI_TOOLS_GOPATH)" \
	GOBIN="$(CRI_TOOLS_GOPATH)/bin" \
	PATH=$(CRI_TOOLS_GOPATH)/bin:$(BR_PATH)


define CRI_TOOLS_CONFIGURE_CMDS
	mkdir -p $(CRI_TOOLS_GOPATH)/src/github.com/kubernetes-incubator
	ln -sf $(@D) $(CRI_TOOLS_GOPATH)/src/github.com/kubernetes-incubator/cri-tools
endef

define CRI_TOOLS_BUILD_CMDS
	$(CRI_TOOLS_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) crictl
endef

define CRI_TOOLS_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
		$(CRI_TOOLS_GOPATH)/bin/crictl \
		$(TARGET_DIR)/usr/bin/crictl
endef

$(eval $(generic-package))
