FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && \
    apt install -y \
        ranger \
        openssh-server \
        nano \
        curl \
        wget \
        iputils-ping \
        net-tools \
        procps && \
    mkdir /var/run/sshd

EXPOSE 22

CMD ["/usr/sbin/sshd","-D"]