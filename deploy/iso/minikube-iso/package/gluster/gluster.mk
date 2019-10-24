################################################################################
#
# gluster
#
################################################################################

GLUSTER_VERSION = 4.1.5
# Official gluster site has SSL problems
# https://bugzilla.redhat.com/show_bug.cgi?id=1572944
# GLUSTER_SITE = https://download.gluster.org/pub/gluster/glusterfs/4.1/$(GLUSTER_VERSION)
GLUSTER_SITE =  http://download.openpkg.org/components/cache/glusterfs
GLUSTER_SOURCE = glusterfs-$(GLUSTER_VERSION).tar.gz
GLUSTER_CONF_OPTS = --disable-tiering --disable-ec-dynamic --disable-xmltest --disable-crypt-xlator --disable-georeplication --disable-ibverbs --disable-glupy --disable-gnfs --disable-cmocka --without-server
GLUSTER_INSTALL_TARGET_OPTS = DESTDIR=$(TARGET_DIR) install
$(eval $(autotools-package))
