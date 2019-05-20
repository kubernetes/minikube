################################################################################
#
# cri-o
#
################################################################################

CRIO_BIN_VERSION = v1.14.1
CRIO_BIN_COMMIT = b7644f67e6383cc862b3e37fb74fba334b0b2721
CRIO_BIN_SITE = https://github.com/kubernetes-sigs/cri-o/archive
CRIO_BIN_SOURCE = $(CRIO_BIN_VERSION).tar.gz
CRIO_BIN_DEPENDENCIES = host-go libgpgme
CRIO_BIN_GOPATH = $(@D)/_output
CRIO_BIN_ENV = \
	CGO_ENABLED=1 \
	GOPATH="$(CRIO_BIN_GOPATH)" \
	GOBIN="$(CRIO_BIN_GOPATH)/bin" \
	PATH=$(CRIO_BIN_GOPATH)/bin:$(BR_PATH)


define CRIO_BIN_USERS
	- -1 crio-admin -1 - - - - -
	- -1 crio       -1 - - - - -
endef

define CRIO_BIN_CONFIGURE_CMDS
	mkdir -p $(CRIO_BIN_GOPATH)/src/github.com/kubernetes-sigs
	ln -sf $(@D) $(CRIO_BIN_GOPATH)/src/github.com/kubernetes-sigs/cri-o
endef

define CRIO_BIN_BUILD_CMDS
	mkdir -p $(@D)/bin
	$(CRIO_BIN_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) GIT_COMMIT=$(CRIO_BIN_COMMIT) PREFIX=/usr binaries
endef

define CRIO_BIN_INSTALL_TARGET_CMDS
	mkdir -p $(TARGET_DIR)/usr/share/containers/oci/hooks.d
	mkdir -p $(TARGET_DIR)/etc/containers/oci/hooks.d

	$(INSTALL) -Dm755 \
		$(@D)/bin/crio \
		$(TARGET_DIR)/usr/bin/crio
	$(INSTALL) -Dm755 \
		$(@D)/bin/conmon \
		$(TARGET_DIR)/usr/libexec/crio/conmon
	$(INSTALL) -Dm755 \
		$(@D)/bin/pause \
		$(TARGET_DIR)/usr/libexec/crio/pause
	$(INSTALL) -Dm644 \
		$(@D)/seccomp.json \
		$(TARGET_DIR)/etc/crio/seccomp.json
	$(INSTALL) -Dm644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/crio-bin/crio.conf \
		$(TARGET_DIR)/etc/crio/crio.conf
	$(INSTALL) -Dm644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/crio-bin/policy.json \
		$(TARGET_DIR)/etc/containers/policy.json
	$(INSTALL) -Dm644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/crio-bin/registries.conf \
		$(TARGET_DIR)/etc/containers/registries.conf

	mkdir -p $(TARGET_DIR)/etc/sysconfig
	echo 'CRIO_OPTIONS="--log-level=debug"' > $(TARGET_DIR)/etc/sysconfig/crio
endef

define CRIO_BIN_INSTALL_INIT_SYSTEMD
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) install.systemd DESTDIR=$(TARGET_DIR) PREFIX=$(TARGET_DIR)/usr
	$(INSTALL) -Dm644 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/crio-bin/crio.service \
		$(TARGET_DIR)/usr/lib/systemd/system/crio.service
	$(call link-service,crio.service)
	$(call link-service,crio-shutdown.service)
endef

$(eval $(generic-package))
