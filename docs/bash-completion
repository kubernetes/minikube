
# Copyright 2016 The Kubernetes Authors.
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
# bash completion for minikube                             -*- shell-script -*-

__debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__my_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__handle_reply()
{
    __debug "${FUNCNAME[0]}"
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            COMPREPLY=( $(compgen -W "${allflags[*]}" -- "$cur") )
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%%=*}"
                __index_of_word "${flag}" "${flags_with_completion[@]}"
                if [[ ${index} -ge 0 ]]; then
                    COMPREPLY=()
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION}" ]; then
                        # zfs completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi
            return 0;
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions=("${must_have_one_noun[@]}")
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    COMPREPLY=( $(compgen -W "${completions[*]}" -- "$cur") )

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        COMPREPLY=( $(compgen -W "${noun_aliases[*]}" -- "$cur") )
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        declare -F __custom_func >/dev/null && __custom_func
    fi

    __ltrim_colon_completions "$cur"
}

# The arguments should be in the form "ext1|ext2|extn"
__handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1
}

__handle_flag()
{
    __debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    if [ -n "${flagvalue}" ] ; then
        flaghash[${flagname}]=${flagvalue}
    elif [ -n "${words[ $((c+1)) ]}" ] ; then
        flaghash[${flagname}]=${words[ $((c+1)) ]}
    else
        flaghash[${flagname}]="true" # pad "true" for bool flag
    fi

    # skip the argument to a two word flag
    if __contains_word "${words[c]}" "${two_word_flags[@]}"; then
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__handle_noun()
{
    __debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__handle_command()
{
    __debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_$(basename "${words[c]//:/__}")"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F $next_command >/dev/null && $next_command
}

__handle_word()
{
    if [[ $c -ge $cword ]]; then
        __handle_reply
        return
    fi
    __debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __handle_flag
    elif __contains_word "${words[c]}" "${commands[@]}"; then
        __handle_command
    elif [[ $c -eq 0 ]] && __contains_word "$(basename "${words[c]}")" "${commands[@]}"; then
        __handle_command
    else
        __handle_noun
    fi
    __handle_word
}

_minikube_addons_disable()
{
    last_command="minikube_addons_disable"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_addons_enable()
{
    last_command="minikube_addons_enable"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_addons_list()
{
    last_command="minikube_addons_list"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_addons_open()
{
    last_command="minikube_addons_open"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    flags+=("--https")
    local_nonpersistent_flags+=("--https")
    flags+=("--url")
    local_nonpersistent_flags+=("--url")
    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_addons()
{
    last_command="minikube_addons"
    commands=()
    commands+=("disable")
    commands+=("enable")
    commands+=("list")
    commands+=("open")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    local_nonpersistent_flags+=("--format=")
    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_completion()
{
    last_command="minikube_completion"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_config_get()
{
    last_command="minikube_config_get"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_config_set()
{
    last_command="minikube_config_set"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_config_unset()
{
    last_command="minikube_config_unset"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_config_view()
{
    last_command="minikube_config_view"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    local_nonpersistent_flags+=("--format=")
    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_config()
{
    last_command="minikube_config"
    commands=()
    commands+=("get")
    commands+=("set")
    commands+=("unset")
    commands+=("view")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_dashboard()
{
    last_command="minikube_dashboard"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--url")
    local_nonpersistent_flags+=("--url")
    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_delete()
{
    last_command="minikube_delete"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_docker-env()
{
    last_command="minikube_docker-env"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--no-proxy")
    local_nonpersistent_flags+=("--no-proxy")
    flags+=("--shell=")
    local_nonpersistent_flags+=("--shell=")
    flags+=("--unset")
    flags+=("-u")
    local_nonpersistent_flags+=("--unset")
    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_get-k8s-versions()
{
    last_command="minikube_get-k8s-versions"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_ip()
{
    last_command="minikube_ip"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_logs()
{
    last_command="minikube_logs"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_mount()
{
    last_command="minikube_mount"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_service_list()
{
    last_command="minikube_service_list"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--namespace=")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--namespace=")
    flags+=("--alsologtostderr")
    flags+=("--format=")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_service()
{
    last_command="minikube_service"
    commands=()
    commands+=("list")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    flags+=("--https")
    local_nonpersistent_flags+=("--https")
    flags+=("--namespace=")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--namespace=")
    flags+=("--url")
    local_nonpersistent_flags+=("--url")
    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_ssh()
{
    last_command="minikube_ssh"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_start()
{
    last_command="minikube_start"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--container-runtime=")
    local_nonpersistent_flags+=("--container-runtime=")
    flags+=("--cpus=")
    local_nonpersistent_flags+=("--cpus=")
    flags+=("--disk-size=")
    local_nonpersistent_flags+=("--disk-size=")
    flags+=("--docker-env=")
    local_nonpersistent_flags+=("--docker-env=")
    flags+=("--extra-config=")
    local_nonpersistent_flags+=("--extra-config=")
    flags+=("--feature-gates=")
    local_nonpersistent_flags+=("--feature-gates=")
    flags+=("--host-only-cidr=")
    local_nonpersistent_flags+=("--host-only-cidr=")
    flags+=("--hyperv-virtual-switch=")
    local_nonpersistent_flags+=("--hyperv-virtual-switch=")
    flags+=("--insecure-registry=")
    local_nonpersistent_flags+=("--insecure-registry=")
    flags+=("--iso-url=")
    local_nonpersistent_flags+=("--iso-url=")
    flags+=("--keep-context")
    local_nonpersistent_flags+=("--keep-context")
    flags+=("--kubernetes-version=")
    local_nonpersistent_flags+=("--kubernetes-version=")
    flags+=("--kvm-network=")
    local_nonpersistent_flags+=("--kvm-network=")
    flags+=("--memory=")
    local_nonpersistent_flags+=("--memory=")
    flags+=("--network-plugin=")
    local_nonpersistent_flags+=("--network-plugin=")
    flags+=("--registry-mirror=")
    local_nonpersistent_flags+=("--registry-mirror=")
    flags+=("--vm-driver=")
    local_nonpersistent_flags+=("--vm-driver=")
    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_status()
{
    last_command="minikube_status"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--format=")
    local_nonpersistent_flags+=("--format=")
    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_stop()
{
    last_command="minikube_stop"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube_version()
{
    last_command="minikube_version"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_minikube()
{
    last_command="minikube"
    commands=()
    commands+=("addons")
    commands+=("completion")
    commands+=("config")
    commands+=("dashboard")
    commands+=("delete")
    commands+=("docker-env")
    commands+=("get-k8s-versions")
    commands+=("ip")
    commands+=("logs")
    commands+=("mount")
    commands+=("service")
    commands+=("ssh")
    commands+=("start")
    commands+=("status")
    commands+=("stop")
    commands+=("version")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--logtostderr")
    flags+=("--show-libmachine-logs")
    flags+=("--stderrthreshold=")
    flags+=("--use-vendored-driver")
    flags+=("--v=")
    two_word_flags+=("-v")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_minikube()
{
    local cur prev words cword
    declare -A flaghash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __my_init_completion -n "=" || return
    fi

    local c=0
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("minikube")
    local must_have_one_flag=()
    local must_have_one_noun=()
    local last_command
    local nouns=()

    __handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_minikube minikube
else
    complete -o default -o nospace -F __start_minikube minikube
fi

# ex: ts=4 sw=4 et filetype=sh
