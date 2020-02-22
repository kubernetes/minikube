VARLINK_VERSION = 18
VARLINK_SITE = https://github.com/varlink/libvarlink/archive
VARLINK_SOURCE = $(VARLINK_VERSION).tar.gz
VARLINK_LICENSE = Apache-2.0
VARLINK_LICENSE_FILES = LICENSE

VARLINK_NEEDS_HOST_PYTHON = python3

define VARLINK_ENV_PYTHON3
    sed -e 's|/usr/bin/python3|/usr/bin/env python3|' -i $(@D)/varlink-wrapper.py
endef

VARLINK_POST_EXTRACT_HOOKS += VARLINK_ENV_PYTHON3

$(eval $(meson-package))
