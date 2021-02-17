################################################################################
#
# cri-o
#
################################################################################

CRIO_BIN_VERSION = v1.20.0
CRIO_BIN_COMMIT = d388528dbed26b93c5bc1c89623607a1e597aa57
CRIO_BIN_SITE = https://github.com/cri-o/cri-o/archive
CRIO_BIN_SOURCE = $(CRIO_BIN_VERSION).tar.gz
CRIO_BIN_DEPENDENCIES = host-go libgpgme
CRIO_BIN_GOPATH = $(@D)/_output
CRIO_BIN_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=1 \
	GO111MODULE=off \
	GOPATH="$(CRIO_BIN_GOPATH)" \
	GOBIN="$(CRIO_BIN_GOPATH)/bin" \
	PATH=$(CRIO_BIN_GOPATH)/bin:$(BR_PATH)


define CRIO_BIN_USERS
	- -1 crio-admin -1 - - - - -
	- -1 crio       -1 - - - - -
endef

define CRIO_BIN_CONFIGURE_CMDS
	mkdir -p $(CRIO_BIN_GOPATH)/src/github.com/cri-o
	ln -sf $(@D) $(CRIO_BIN_GOPATH)/src/github.com/cri-o/cri-o
	# disable the "automatic" go module detection
	sed -e 's/go help mod/false/' -i $(@D)/Makefile
endef

define CRIO_BIN_BUILD_CMDS
	mkdir -p $(@D)/bin
	$(CRIO_BIN_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) COMMIT_NO=$(CRIO_BIN_COMMIT) PREFIX=/usr binaries
endef

define CRIO_BIN_INSTALL_TARGET_CMDS
	mkdir -p $(TARGET_DIR)/usr/share/containers/oci/hooks.d
	mkdir -p $(TARGET_DIR)/etc/containers/oci/hooks.d

	$(INSTALL) -Dm755 \
		$(@D)/bin/crio \
		$(TARGET_DIR)/usr/bin/crio
	$(INSTALL) -Dm755 \
		$(@D)/bin/pinns \
		$(TARGET_DIR)/usr/bin/pinns
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_PKGDIR)/crio.conf \
		$(TARGET_DIR)/etc/crio/crio.conf
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_PKGDIR)/policy.json \
		$(TARGET_DIR)/etc/containers/policy.json
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_PKGDIR)/registries.conf \
		$(TARGET_DIR)/etc/containers/registries.conf

	mkdir -p $(TARGET_DIR)/etc/sysconfig
	echo 'CRIO_OPTIONS="--log-level=debug"' > $(TARGET_DIR)/etc/sysconfig/crio
endef

define CRIO_BIN_INSTALL_INIT_SYSTEMD
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) install.systemd DESTDIR=$(TARGET_DIR) PREFIX=$(TARGET_DIR)/usr
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_PKGDIR)/crio.service \
		$(TARGET_DIR)/usr/lib/systemd/system/crio.service
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_PKGDIR)/crio-wipe.service \
		$(TARGET_DIR)/usr/lib/systemd/system/crio-wipe.service
	$(call link-service,crio.service)
	$(call link-service,crio-shutdown.service)
endef

$(eval $(generic-package))
