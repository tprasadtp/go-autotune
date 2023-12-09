#!/bin/bash

#!/usr/bin/env bash
#  Copyright (c) 2023, Prasad Tengse
# shellcheck disable=SC2034,SC2155

set -o pipefail

# Script Constants
readonly CURDIR="$(cd -P -- "$(dirname -- "")" && pwd -P)"
readonly SCRIPT="$(basename "$0")"

# Handle Signals
# trap ctrl-c and SIGTERM
trap ctrl_c_signal_handler INT
trap term_signal_handler SIGTERM

function ctrl_c_signal_handler() {
    log_error "User Interrupt! CTRL-C"
    exit 4
}

function term_signal_handler() {
    log_error "Signal Interrupt! SIGTERM"
    exit 4
}

#diana::snippet:shlib-logger:begin#
function __is_stderr_colorable() {
    # CLICOLOR_FORCE is set and CLICOLOR_FORCE != 0, force colors
    if [[ -n ${CLICOLOR_FORCE} ]] && [[ ${CLICOLOR_FORCE} != "0" ]]; then
        return 0

    # CLICOLOR == 0 or NO_COLOR is set and not empty or TERM is dumb or linux
    elif [[ -n ${NO_COLOR} ]] || [[ ${CLICOLOR} == "0" ]] || [[ ${TERM} == "dumb" ]] || [[ ${TERM} == "linux" ]]; then
        return 1
    fi

    if [[ -t 2 ]]; then
        return 0
    fi
    return 1
}

# Logger core ::internal::
function __logger_core_event_handler() {
    [[ $# -lt 2 ]] && return 1

    local lvl_caller="${1:-info}"
    case ${lvl_caller} in
    trace)
        level="0"
        ;;
    debug)
        level="10"
        ;;
    info)
        level="20"
        ;;
    success)
        level="20"
        ;;
    notice)
        level="25"
        ;;
    warning)
        level="30"
        ;;
    error)
        level="40"
        ;;
    *)
        level="100"
        ;;
    esac

    # Immediately return if log level is not enabled, If LOG_LVL is not set, defaults to 20 - info level
    if [[ ${LOG_LVL:-20} -gt "${level}" ]]; then
        return
    fi

    shift
    local lvl_msg="$*"

    local lvl_color
    local lvl_colorized
    local lvl_reset

    if __is_stderr_colorable; then
        lvl_colorized="true"
        # shellcheck disable=SC2155
        lvl_reset="\e[0m"
    fi

    # Level name in string format
    local lvl_prefix
    # Level name in string format with timestamp if enabled or level symbol
    local lvl_string

    # Log format
    if [[ ${LOG_FMT:-pretty} == "pretty" ]] && [[ -n ${lvl_colorized} ]]; then
        lvl_string="[•]"
    elif [[ ${LOG_FMT} = "full" ]] || [[ ${LOG_FMT} = "long" ]]; then
        if [[ ${LOG_LVL:-20} -lt 20 ]]; then
            printf -v lvl_prefix "%(%FT%TZ)T (%-4s) " -1 "${BASH_LINENO[1]}"
        else
            printf -v lvl_prefix "%(%FT%TZ)T" -1
        fi
    elif [[ ${LOG_FMT} = "journald" ]] || [[ ${LOG_FMT} = "journal" ]]; then
        if [[ ${LOG_LVL:-20} -lt 20 ]]; then
            printf -v lvl_prefix "(%-4s) " "${BASH_LINENO[1]}"
        fi
    fi

    # Define level, color and timestamp
    # By default we do not show log level and timestamp.
    # However, if LOG_FMT is set to "full" or "long", we will enable long format with timestamps
    case "$lvl_caller" in
    trace)
        # if lvl_string is set earlier, that means LOG_FMT is default or pretty
        # we dont display timestamp or level name in this case. otherwise
        # append level name to lvl_prefix
        # (lvl_prefix is populated with timestamp if LOG_FMT is full or long)
        [[ -z ${lvl_string} ]] && lvl_string="${lvl_prefix}[TRACE   ]"
        [[ -n "${lvl_colorized}" ]] && lvl_color="\e[38;5;246m"
        ;;
    debug)
        [[ -z ${lvl_string} ]] && lvl_string="${lvl_prefix}[DEBUG   ]"
        [[ -n "${lvl_colorized}" ]] && lvl_color="\e[38;5;250m"
        ;;
    info)
        [[ -z ${lvl_string} ]] && lvl_string="${lvl_prefix}[INFO    ]"
        # Avoid printing color reset sequence as this level is not colored
        [[ -n "${lvl_colorized}" ]] && lvl_reset=""
        ;;
    success)
        [[ -z ${lvl_string} ]] && lvl_string="${lvl_prefix}[SUCCESS ]"
        [[ -n "${lvl_colorized}" ]] && lvl_color="\e[38;5;83m"
        ;;
    notice)
        [[ -z ${lvl_string} ]] && lvl_string="${lvl_prefix}[NOTICE  ]"
        # shellcheck disable=SC2155
        [[ -n "${lvl_colorized}" ]] && lvl_color="\e[38;5;81m"
        ;;
    warning)
        [[ -z ${lvl_string} ]] && lvl_string="${lvl_prefix}[WARNING ]"
        # shellcheck disable=SC2155
        [[ -n "${lvl_colorized}" ]] && lvl_color="\e[38;5;214m"
        ;;
    error)
        [[ -z ${lvl_string} ]] && lvl_string="${lvl_prefix}[ERROR   ]"
        # shellcheck disable=SC2155
        [[ -n "${lvl_colorized}" ]] && lvl_color="\e[38;5;197m"
        ;;
    *)
        [[ -z ${lvl_string} ]] && lvl_string="${lvl_prefix}[UNKNOWN ]"
        # Avoid printing color reset sequence as this level is not colored
        [[ -n "${lvl_colorized}" ]] && lvl_reset=""
        ;;
    esac

    printf "${lvl_color}%s %s ${lvl_reset}\n" "${lvl_string}" "$lvl_msg"
}

function log_trace() {
    __logger_core_event_handler "trace" "$@"
}

function log_debug() {
    __logger_core_event_handler "debug" "$@"
}

function log_info() {
    __logger_core_event_handler "info" "$@"
}

function log_success() {
    __logger_core_event_handler "success" "$@"
}

function log_warning() {
    __logger_core_event_handler "warning" "$@"
}

function log_notice() {
    __logger_core_event_handler "notice" "$@"
}

function log_error() {
    __logger_core_event_handler "error" "$@"
}

function log_variable() {
    local var="$1"
    local __msg_string
    printf -v __msg_string "%s : %s" "${var}" "${!var:-NA}"
    __logger_core_event_handler "debug" "${__msg_string}"
}

function log_kv_pair() {
    local __msg_string
    printf -v __msg_string "%-${4:-25}s : %s" "${1:-NA}" "${2:-NA}"
    __logger_core_event_handler "debug" "${__msg_string}"
}

function log_tail() {
    local line prefix
    [[ -n $1 ]] && prefix="($1) "
    while read -r line; do
        __logger_core_event_handler "trace" "$prefix$line"
    done
}
#diana::snippet:shlib-logger:end#

# Checks if command is available
function has_command() {
    if command -v "$1" >/dev/null; then
        return 0
    else
        return 1
    fi
    return 1
}

function display_usage() {
    cat <<EOF
Script to generate cgroup and procfs data

Usage: ${SCRIPT} [OPTIONS]...

Arguments:
  None

Options:
  -h, --help          Display this help message
  -v, --verbose       Increase log verbosity

Examples:
  ${SCRIPT} --help    Display help

Environment:
  NO_COLOR            Set this to NON-EMPTY to disable all colors.
  CLICOLOR_FORCE      Set this to NON-ZERO to force colored output.
EOF
}

function main() {
    while [[ ${1} != "" ]]; do
        case ${1} in
        # Debugging options
        --stdout) LOG_TO_STDOUT="true" ;;
        -v | --verbose)
            LOG_LVL="1"
            log_info "Enable verbose logging"
            ;;
        -h | --help)
            display_usage
            exit 0
            ;;
        *)
            log_error "Invalid argument(s). See usage below."
            display_usage
            exit 1
            ;;
        esac
        shift
    done

    if ! has_command systemd-run; then
        log_error "Missing command systemd-run"
        exit 1
    else
        log_debug "Found command systemd-run"
    fi

    if ! has_command rsync; then
        log_error "Missing command rsync"
        exit 1
    else
        log_debug "Found command rsync"
    fi

    if ! has_command docker; then
        log_error "Missing command docker"
        exit 1
    else
        log_debug "Found command docker"
    fi

    if [[ ! -d /run/systemd/ ]]; then
        log_error "Not booted using systemd"
    else
        log_debug "System is booted using systmed"
    fi

    SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
    log_variable "SCRIPT_DIR"

    log_info "Generating test files for no limits specified (systmed)"


}

main "$@"
