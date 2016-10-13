################################################################################
#
# rkt
#
################################################################################

RKT_BIN_VERSION = 1.14.0
RKT_BIN_SITE = https://github.com/coreos/rkt/releases/download/v$(RKT_BIN_VERSION)
RKT_BIN_SOURCE = rkt-v$(RKT_BIN_VERSION).tar.gz

RKT_BIN_EXTRA_DOWNLOADS = \
	https://github.com/coreos/rkt/releases/download/v$(RKT_BIN_VERSION)/rkt-v$(RKT_BIN_VERSION).tar.gz.asc \
	https://coreos.com/dist/pubkeys/app-signing-pubkey.gpg

define RKT_BIN_USERS
	- -1 rkt-admin -1 - - - - -
	- -1 rkt       -1 - - - - -
endef

define RKT_BIN_BUILD_CMDS
	gpg2 --import $(BR2_DL_DIR)/app-signing-pubkey.gpg

	gpg2 \
		--trusted-key $(shell gpg2 --with-colons --keyid-format LONG -k security@coreos.com | egrep ^pub | cut -d ':' -f5) \
		--verify-files $(BR2_DL_DIR)/rkt-v$(RKT_BIN_VERSION).tar.gz.asc

	mkdir -p $(TARGET_DIR)/var/lib/rkt
endef

define RKT_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/rkt \
		$(TARGET_DIR)/bin/rkt

	mkdir -p $(TARGET_DIR)/etc/bash_completion.d

	$(INSTALL) -D -m 644 \
		$(@D)/bash_completion/rkt.bash \
		$(TARGET_DIR)/etc/bash_completion.d/rkt

	mkdir -p $(TARGET_DIR)/usr/lib/rkt/stage1-images

	install -Dm644 \
		$(@D)/stage1-coreos.aci \
		$(TARGET_DIR)/usr/lib/rkt/stage1-images/stage1-coreos.aci
endef

define RKT_BIN_INSTALL_INIT_SYSTEMD
	mkdir -p $(TARGET_DIR)/usr/lib/tmpfiles.d

	$(INSTALL) -D -m 644 \
		$(@D)/init/systemd/tmpfiles.d/rkt.conf \
		$(TARGET_DIR)/usr/lib/tmpfiles.d/rkt.conf

	$(call install-service,rkt-api.service)
	$(call install-service,rkt-gc.timer)
	$(call install-service,rkt-gc.service)
	$(call install-service,rkt-metadata.socket)
	$(call install-service,rkt-metadata.service)
endef

define install-service
	$(INSTALL) -D -m 644 \
		$(@D)/init/systemd/$(1) \
		$(TARGET_DIR)/usr/lib/systemd/system/$(1)

	ln -fs /usr/lib/systemd/system/$(1) \
		$(TARGET_DIR)/etc/systemd/system/multi-user.target.wants/$(1)
endef

$(eval $(generic-package))
