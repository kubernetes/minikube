VARLINK_VERSION = 18
VARLINK_SITE = https://github.com/varlink/libvarlink/archive
VARLINK_SOURCE = $(VARLINK_VERSION).tar.gz
VARLINK_LICENSE = Apache-2.0
VARLINK_LICENSE_FILES = LICENSE

$(eval $(meson-package))
