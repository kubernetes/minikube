################################################################################
#
# cri-o
#
################################################################################

CRIO_BIN_VERSION = v1.0.0
CRIO_BIN_SITE = https://github.com/kubernetes-incubator/cri-o/archive
CRIO_BIN_SOURCE = $(CRIO_BIN_VERSION).tar.gz
CRIO_BIN_DEPENDENCIES = libgpgme
CRIO_BIN_GOPATH = $(@D)/_output
CRIO_BIN_ENV = \
	GOPATH="$(CRIO_BIN_GOPATH)" \
	PATH=$(CRIO_BIN_GOPATH)/bin:$(BR_PATH)


define CRIO_BIN_USERS
	- -1 crio-admin -1 - - - - -
	- -1 crio       -1 - - - - -
endef

define CRIO_BIN_CONFIGURE_CMDS
	mkdir -p $(CRIO_BIN_GOPATH)/src/github.com/kubernetes-incubator
	ln -sf $(@D) $(CRIO_BIN_GOPATH)/src/github.com/kubernetes-incubator/cri-o
	$(CRIO_BIN_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) install.tools DESTDIR=$(TARGET_DIR) PREFIX=$(TARGET_DIR)/usr
endef

define CRIO_BIN_BUILD_CMDS
	$(CRIO_BIN_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) binaries PREFIX=/usr BUILDTAGS="containers_image_ostree_stub" 
endef

define CRIO_BIN_INSTALL_TARGET_CMDS
	mkdir -p $(TARGET_DIR)/usr/share/containers/oci/hooks.d
	mkdir -p $(TARGET_DIR)/etc/containers/oci/hooks.d

	$(INSTALL) -Dm755 \
		$(@D)/crio \
		$(TARGET_DIR)/usr/bin/crio
	$(INSTALL) -Dm755 \
		$(@D)/crioctl \
		$(TARGET_DIR)/usr/bin/crioctl
	$(INSTALL) -Dm755 \
		$(@D)/kpod \
		$(TARGET_DIR)/usr/bin/kpod
	$(INSTALL) -Dm755 \
		$(@D)/conmon/conmon \
		$(TARGET_DIR)/usr/libexec/crio/conmon
	$(INSTALL) -Dm755 \
		$(@D)/pause/pause \
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

	mkdir -p $(TARGET_DIR)/etc/sysconfig
	echo 'CRIO_OPTIONS="--storage-driver=overlay2 --log-level=debug"' > $(TARGET_DIR)/etc/sysconfig/crio
endef

define CRIO_BIN_INSTALL_INIT_SYSTEMD
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) install.systemd DESTDIR=$(TARGET_DIR) PREFIX=$(TARGET_DIR)/usr
	$(INSTALL) -Dm755 \
		$(BR2_EXTERNAL_MINIKUBE_PATH)/package/crio-bin/crio.service \
		$(TARGET_DIR)/usr/lib/systemd/system/crio.service
	$(call link-service,crio.service)
	$(call link-service,crio-shutdown.service)
endef

$(eval $(generic-package))
