# Copyright 2016 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# TODO: Move this to somewhere on GCR and manage tags explicitly instead of
# relying on latest.
FROM boot2docker/boot2docker

# Set the version of Docker and the expected sha.
RUN echo "1.11.1" > $ROOTFS/etc/version
ENV DOCKER_SHA b7d6025beb6440341741ea9324655a267fba83f7

# Add other dependencies here to $ROOTFS/
ADD nsenter $ROOTFS/usr/bin/
ADD socat $ROOTFS/usr/bin
ADD ethtool $ROOTFS/usr/bin
ADD conntrack $ROOTFS/usr/bin
ADD bootlocal.sh $ROOTFS/tmp/bootlocal.sh
RUN cat $ROOTFS/tmp/bootlocal.sh >> $ROOTFS/etc/rc.d/automount

# Get a specific version of Docker. This will overwrite the binary installed
# in the base image.
# The --strip-components=3 flag will have to change as we switch docker versions
# past 1.10.x. They changed the packaging format slightly.
RUN rm -f /$ROOTFS/usr/local/bin/docker*
RUN curl -fSL -o /tmp/dockerbin.tgz https://get.docker.com/builds/Linux/x86_64/docker-$(cat $ROOTFS/etc/version).tgz && \
	# Check the sha1 matches.
	if [ $DOCKER_SHA != $(sha1sum /tmp/dockerbin.tgz | cut -f1 -d ' ') ]; then echo "SHA mismatch!"; exit 1; fi && \
    tar -zxvf /tmp/dockerbin.tgz -C "$ROOTFS/usr/local/bin" --strip-components=1 && \
    rm /tmp/dockerbin.tgz

RUN $ROOTFS/usr/local/bin/docker -v

# Build the actual iso image
RUN /make_iso.sh

# Output it to transfer back to the host
CMD ["cat", "boot2docker.iso"]
