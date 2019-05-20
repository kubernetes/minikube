################################################################################
#
# containerd
#
################################################################################
CONTAINERD_BIN_VERSION = v1.2.5
CONTAINERD_BIN_COMMIT = bb71b10fd8f58240ca47fbb579b9d1028eea7c84
CONTAINERD_BIN_SITE = https://github.com/containerd/containerd/archive
CONTAINERD_BIN_SOURCE = $(CONTAINERD_BIN_VERSION).tar.gz
CONTAINERD_BIN_DEPENDENCIES = host-go libgpgme
CONTAINERD_BIN_GOPATH = $(@D)/_output
CONTAINERD_BIN_ENV = \
	CGO_ENABLED=1 \
	GOPATH="$(CONTAINERD_BIN_GOPATH)" \
	GOBIN="$(CONTAINERD_BIN_GOPATH)/bin" \
	PATH=$(CONTAINERD_BIN_GOPATH)/bin:$(BR_PATH)

CONTAINERD_BIN_COMPILE_SRC = $(CONTAINERD_BIN_GOPATH)/src/github.com/containerd/containerd

define CONTAINERD_BIN_USERS
	- -1 containerd-admin -1 - - - - -
	- -1 containerd       -1 - - - - -
endef

define CONTAINERD_BIN_CONFIGURE_CMDS
	mkdir -p $(CONTAINERD_BIN_GOPATH)/src/github.com/containerd
	ln -sf $(@D) $(CONTAINERD_BIN_COMPILE_SRC)
endef

define CONTAINERD_BIN_BUILD_CMDS
	PWD=$(CONTAINERD_BIN_COMPILE_SRC) $(CONTAINERD_BIN_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) VERSION=$(CONTAINERD_BIN_VERSION) REVISION=$(CONTAINERD_BIN_COMMIT) -C $(@D) binaries
endef

define CONTAINERD_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
		$(@D)/bin/containerd \
		$(TARGET_DIR)/usr/bin
	$(INSTALL) -Dm755 \
		$(@D)/bin/containerd-shim \
		$(TARGET_DIR)/usr/bin
	$(INSTALL) -Dm755 \
		$(@D)/bin/ctr \
		$(TARGET_DIR)/usr/bin
	$(INSTALL) -Dm644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/containerd-bin/config.toml \
		$(TARGET_DIR)/etc/containerd/config.toml
endef

define CONTAINERD_BIN_INSTALL_INIT_SYSTEMD
	$(INSTALL) -Dm755 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/containerd-bin/containerd.service \
		$(TARGET_DIR)/usr/lib/systemd/system/containerd.service
	$(call link-service,containerd.service)
	$(call link-service,containerd-shutdown.service)
endef

$(eval $(generic-package))
