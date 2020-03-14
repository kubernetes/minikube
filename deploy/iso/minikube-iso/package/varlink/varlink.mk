VARLINK_VERSION = 19
VARLINK_SITE = https://github.com/varlink/libvarlink/archive
VARLINK_SOURCE = $(VARLINK_VERSION).tar.gz
VARLINK_LICENSE = Apache-2.0
VARLINK_LICENSE_FILES = LICENSE

VARLINK_NEEDS_HOST_PYTHON = python3

$(eval $(meson-package))
