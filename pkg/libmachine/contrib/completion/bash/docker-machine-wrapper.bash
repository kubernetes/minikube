#
# Function wrapper to docker-machine that adds a use subcommand.
#
# The use subcommand runs `eval "$(docker-machine env [args])"`, which is a lot
# less typing.
#
# To enable:
#  1a. Copy this file somewhere and source it in your .bashrc
#      source /some/where/docker-machine-wrapper.bash
#  1b. Alternatively, just copy this file into into /etc/bash_completion.d
#
# Configuration:
#
# DOCKER_MACHINE_WRAPPED
#   When set to a value other than true, this will disable the alias wrapper
#   alias for docker-machine. This is useful if you don't want the wrapper,
#   but it is installed by default by your installation.
#

: ${DOCKER_MACHINE_WRAPPED:=true}

__docker_machine_wrapper () {
    if [[ "$1" == use ]]; then
        # Special use wrapper
        shift 1
        case "$1" in
            -h|--help|"")
                cat <<EOF
Usage: docker-machine use [OPTIONS] [arg...]

Evaluate the commands to set up the environment for the Docker client

Description:
   Argument is a machine name.

Options:

   --swarm	Display the Swarm config instead of the Docker daemon
   --unset, -u	Unset variables instead of setting them

EOF
                ;;
            *)
                eval "$(docker-machine env "$@")"
                echo "Active machine: ${DOCKER_MACHINE_NAME}"
                ;;
        esac
    else
        # Just call the actual docker-machine app
        command docker-machine "$@"
    fi
}

if [[ ${DOCKER_MACHINE_WRAPPED} = true ]]; then
    alias docker-machine=__docker_machine_wrapper
fi
