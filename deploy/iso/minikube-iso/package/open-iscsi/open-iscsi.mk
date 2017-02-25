#############################################################
#
# open-iscsi
#
#############################################################

OPENISCSI_VER:=1.0-485
OPENISCSI_SITE:=http://www.open-iscsi.org/bits
OPENISCSI_DIR:=$(BUILD_DIR)/open-iscsi-$(OPENISCSI_VER)
OPENISCSI_SOURCE:=open-iscsi-$(OPENISCSI_VER).tar.gz

$(DL_DIR)/$(OPENISCSI_SOURCE):
	$(WGET) -P $(DL_DIR) $(OPENISCSI_SITE)/$(OPENISCSI_SOURCE)

$(OPENISCSI_DIR)/.unpacked: $(DL_DIR)/$(OPENISCSI_SOURCE)
	zcat $(DL_DIR)/$(OPENISCSI_SOURCE) | tar -C $(BUILD_DIR) $(TAR_OPTIONS)
	cat package/open-iscsi/open-iscsi-$(OPENISCSI_VER).patch | (cd (OPENISCSI_DIR); patch -p1);
	touch $(OPENISCSI_DIR)/.unpacked

BINS = iscsid iscsiadm
TARGET_BINS:=$(patsubst %,$(TARGET_DIR)/sbin/%, $(BINS))

#$(OPENISCSI_DIR)/usr/%: open-iscsi-unpack
#	@echo "Building $(notdir $@)"
#	cd $(OPENISCSI_DIR)/usr; $(MAKE) $(notdir $@) CC=$(TARGET_CC)
#
#$(TARGET_DIR)/sbin/%: $(OPENISCSI_DIR)/usr/%
#	@echo "Installing $(notdir $@)"
#	cp $^ $@

$(TARGET_DIR)/sbin/%: open-iscsi-unpack
	@echo "Building $(notdir $@)"
	cd $(OPENISCSI_DIR)/usr; $(MAKE) $(notdir $@) CC=$(TARGET_CC)
	@echo "Installing $(notdir $@)"
	cp $(OPENISCSI_DIR)/usr/$(notdir $@) $@

open-iscsi: berkeleydb $(TARGET_BINS)

open-iscsi-source: $(DL_DIR)/$(OPENISCSI_SOURCE)

open-iscsi-unpack: $(OPENISCSI_DIR)/.unpacked

open-iscsi-clean: 
	-$(MAKE) -C $(OPENISCSI_DIR) clean
	-rm -f $(TARGET_BINS)

open-iscsi-dirclean: 
	rm -rf $(OPENISCSI_DIR)

#############################################################
#
# Toplevel Makefile options
#
#############################################################
ifeq ($(strip $(BR2_PACKAGE_OPENISCSI)),y)
TARGETS+=open-iscsi
endif
