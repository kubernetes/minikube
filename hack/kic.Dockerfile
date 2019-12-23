ARG COMMIT_SHA
FROM kindest/node:v1.16.2
RUN apt-get update && apt-get install -y \
  sudo \
  dnsutils \
  && rm -rf /var/lib/apt/lists/*
RUN echo "kic! ${TRAVIS_COMMIT}-${KUBE_VER}" > "/kic.txt"
