################################################################################
#
# containerd
#
################################################################################
CONTAINERD_BIN_VERSION = v1.7.23
CONTAINERD_BIN_COMMIT = 57f17b0a6295a39009d861b89e3b3b87b005ca27
CONTAINERD_BIN_SITE = https://github.com/containerd/containerd/archive
CONTAINERD_BIN_SOURCE = $(CONTAINERD_BIN_VERSION).tar.gz
CONTAINERD_BIN_DEPENDENCIES = host-go libgpgme
CONTAINERD_BIN_GOPATH = $(@D)/_output
CONTAINERD_BIN_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=1 \
	GO111MODULE=off \
	GOPATH="$(CONTAINERD_BIN_GOPATH)" \
	GOBIN="$(CONTAINERD_BIN_GOPATH)/bin" \
	PATH=$(CONTAINERD_BIN_GOPATH)/bin:$(BR_PATH) \
	GOARCH=amd64

CONTAINERD_BIN_COMPILE_SRC = $(CONTAINERD_BIN_GOPATH)/src/github.com/containerd/containerd

define CONTAINERD_BIN_USERS
	- -1 containerd-admin -1 - - - - -
	- -1 containerd       -1 - - - - -
endef

define CONTAINERD_BIN_CONFIGURE_CMDS
	mkdir -p $(CONTAINERD_BIN_GOPATH)/src/github.com/containerd
	mkdir -p $(TARGET_DIR)/etc/containerd/containerd.conf.d
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
		$(@D)/bin/containerd-shim-runc-v1 \
		$(TARGET_DIR)/usr/bin
	$(INSTALL) -Dm755 \
		$(@D)/bin/containerd-shim-runc-v2 \
		$(TARGET_DIR)/usr/bin
	$(INSTALL) -Dm755 \
		$(@D)/bin/ctr \
		$(TARGET_DIR)/usr/bin
	$(INSTALL) -Dm644 \
		$(CONTAINERD_BIN_PKGDIR)/config.toml \
		$(TARGET_DIR)/etc/containerd/config.toml
	$(INSTALL) -Dm644 \
		$(CONTAINERD_BIN_PKGDIR)/containerd_docker_io_hosts.toml \
		$(TARGET_DIR)/etc/containerd/certs.d/docker.io/hosts.toml
endef

define CONTAINERD_BIN_INSTALL_INIT_SYSTEMD
	$(INSTALL) -Dm644 \
		$(CONTAINERD_BIN_PKGDIR)/containerd.service \
		$(TARGET_DIR)/usr/lib/systemd/system/containerd.service
	$(INSTALL) -Dm644 \
		$(CONTAINERD_BIN_PKGDIR)/50-minikube.preset \
		$(TARGET_DIR)/usr/lib/systemd/system-preset/50-minikube.preset
endef

$(eval $(generic-package))
