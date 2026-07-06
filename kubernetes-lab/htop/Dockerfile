FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && \
    apt install -y \
        htop \
        nano \
        curl \
        wget \
        iputils-ping \
        net-tools \
        procps && \
    apt clean

CMD ["sleep", "infinity"]