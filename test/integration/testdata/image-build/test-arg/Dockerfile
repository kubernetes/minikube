FROM registry.k8s.io/alpine-with-bash:1.0
ARG ENV_A
ARG ENV_B
RUN echo "test-build-arg" $ENV_A $ENV_B
CMD ["/bin/sh", "-c"]
