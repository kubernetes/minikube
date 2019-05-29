# BUILDAH_NOPIVOT=true disables pivot_root in Buildah, using MS_MOVE instead.
# (Buildah is used by Podman for building container images using a Dockerfile)
export BUILDAH_NOPIVOT=true
