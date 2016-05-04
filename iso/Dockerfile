# TODO: Move this to somewhere on GCR and manage tags explicitly instead of relying on latest.
FROM boot2docker/boot2docker

# Add other dependencies here to $ROOTFS/
ADD nsenter $ROOTFS/usr/bin/

RUN /make_iso.sh
CMD ["cat", "boot2docker.iso"]