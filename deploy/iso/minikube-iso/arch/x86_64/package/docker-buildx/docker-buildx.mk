################################################################################
# 
# docker-buildx
#   
################################################################################

DOCKER_BUILDX_VERSION = v0.17.1
DOCKER_BUILDX_COMMIT = 257815a6fbaee88976808020bf04274388275ae8
DOCKER_BUILDX_SITE = https://github.com/docker/buildx/archive
DOCKER_BUILDX_SOURCE = $(DOCKER_BUILDX_VERSION).tar.gz
DOCKER_BUILDX_GOPATH = $(@D)/_output
DOCKER_BUILDX_ENV = \
        $(GO_TARGET_ENV) \
        CGO_ENABLED=1 \
        GO111MODULE=on \
        GOPATH="$(DOCKER_BUILDX_GOPATH)" \
        GOBIN="$(DOCKER_BUILDX_GOPATH)/bin" \
        PATH=$(DOCKER_BUILDX_GOPATH)/bin:$(BR_PATH) \
        GOARCH=amd64

DOCKER_BUILDX_COMPILE_SRC = $(DOCKER_BUILDX_GOPATH)/src/github.com/docker/buildx

define DOCKER_BUILDX_POST_EXTRACT_WORKAROUNDS
        # Set -buildvcs=false to disable VCS stamping (fails in buildroot)
        sed -i 's|go build |go build -buildvcs=false |' -i $(@D)/hack/build
endef

DOCKER_BUILDX_POST_EXTRACT_HOOKS += DOCKER_BUILDX_POST_EXTRACT_WORKAROUNDS

define DOCKER_BUILDX_CONFIGURE_CMDS
        mkdir -p $(TARGET_DIR)/usr/libexec/docker/cli-plugins
endef

define DOCKER_BUILDX_BUILD_CMDS
        PWD=$(DOCKER_BUILDX_COMPILE_SRC) $(DOCKER_BUILDX_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) VERSION=$(DOCKER_BUILDX_VERSION) REVISION=$(DOCKER_BUILDX_COMMIT) -C $(@D) build
endef

define DOCKER_BUILDX_INSTALL_TARGET_CMDS
	$(INSTALL) -Dm755 \
                $(@D)/bin/build/docker-buildx \
                $(TARGET_DIR)/usr/libexec/docker/cli-plugins/docker-buildx
endef

$(eval $(generic-package))
