# Plain make targets if not requested inside a container
ifneq (,$(findstring test-integration,$(MAKECMDGOALS)))
	include Makefile.inc
	include mk/main.mk
else ifneq ($(USE_CONTAINER), true)
	include Makefile.inc
	include mk/main.mk
else
# Otherwise, with docker, swallow all targets and forward into a container
DOCKER_BUILD_DONE := ""

test: .DEFAULT

.DEFAULT:
	@test ! -z "$(DOCKER_BUILD_DONE)" || ./script/build_in_container.sh $(MAKECMDGOALS)
	$(eval DOCKER_BUILD_DONE := "done")

endif
