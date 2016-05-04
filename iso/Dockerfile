# TODO: Move this to somewhere on GCR and manage tags explicitly instead of
# relying on latest.
FROM boot2docker/boot2docker

# Set the version of Docker.
RUN echo "1.9.1" > $ROOTFS/etc/version

# Add other dependencies here to $ROOTFS/
ADD nsenter $ROOTFS/usr/bin/

# Get a specific version of Docker. This will overwrite the binary installed
# in the base image.
# The --strip-components=3 flag will have to change as we switch docker versions
# past 1.10.x. They changed the packaging format slightly.
RUN rm -f /$ROOTFS/usr/local/bin/docker*
RUN curl -fSL -o /tmp/dockerbin.tgz https://get.docker.com/builds/Linux/x86_64/docker-$(cat $ROOTFS/etc/version).tgz && \
    tar -zxvf /tmp/dockerbin.tgz -C "$ROOTFS/usr/local/bin" --strip-components=3 && \
    rm /tmp/dockerbin.tgz
RUN $ROOTFS/usr/local/bin/docker -v

# Build the actual iso image
RUN /make_iso.sh

# Output it to transfer back to the host
CMD ["cat", "boot2docker.iso"]
