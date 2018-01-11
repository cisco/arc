#
# Copyright (c) 2017, Cisco Systems
# All rights reserved.
#
# Redistribution and use in source and binary forms, with or without modification,
# are permitted provided that the following conditions are met:
#
# * Redistributions of source code must retain the above copyright notice, this
#   list of conditions and the following disclaimer.
#
# * Redistributions in binary form must reproduce the above copyright notice, this
#   list of conditions and the following disclaimer in the documentation and/or
#   other materials provided with the distribution.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
# ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
# WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
# DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
# ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
# (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
# LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
# ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
# (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
# SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
#

# return if already loaded
[ -n "${common_library_sourced:-}" ] && return 0
readonly common_library_sourced='yes'

# Abort script at first error
set -e
set -o pipefail

# Debugging
set +x; if [ "$debug" = "yes" ]; then set -x; fi

# Syntax checking: read commands in script, but do not execute them
set +n; if [ "$syntax_check" = "yes" ]; then set -n; fi

# Disable side effects
declare safe
if [ "$side_effects" = "no" ]; then safe="detail"; fi

function allow_side_effects() {
  if [ "$side_effects" = "no" ]; then return $failure; fi
  return $success
}

# Setup well know paths
#---------------------------------------------------------------

declare -r tmp_dir="/tmp"

# Globals
#---------------------------------------------------------------
declare -ri success=0
declare -ri failure=1

# Colors
#---------------------------------------------------------------
declare -r esc=$'\033'
declare -r title="${esc}[33;32m"
declare -r header="${esc}[33;36m"
declare -r detail="${esc}[33;33m"
declare -r error="${esc}[33;31m"
declare -r warn="${esc}[33;35m"
declare -r clear="${esc}[33;0m"

# Messaging
#---------------------------------------------------------------
declare -r indent="  "

function error() {
  printf "${error}Error:${clear} $@\n\n" 1>&2
}

function warn() {
  printf "${warn}Warning:${clear} $@\n\n"
}

function info() {
  printf "${title}$@${clear}\n"
}

function verbose() {
  if [ "$verbose" = "yes" ]; then
    printf "${indent}$@\n"
  fi
}

function detail() {
  local m=$(echo $@)
  printf "${indent}${detail}$(echo $@)${clear}\n"
}

function die() {
  if [ $# -gt 0 ]; then error "$@"; fi
  exit 1
}

# String manipulations
#---------------------------------------------------------------

function lower() {
  echo $1 | tr [:upper:] [:lower:]
}

function capitalize() {
  echo $(echo -n ${1:0:1} | tr [:lower:] [:upper:])$(echo ${1:1})
}

# Temp files
#---------------------------------------------------------------

function create_tmp_file() {
  local tmp_file
  local result

  tmp_file=$(mktemp $tmp_dir/XXXXXXXXXX); result=$?
  trap "rm -f $tmp_file" EXIT SIGINT SIGTERM

  if [ $result -ne $success ]; then
    error "Failed to create tmp file" || return $failure
  fi
  echo $tmp_file
}

function create_tmp_dir() {
  local _tmp_dir
  local result

  if [ "$1" = "" ]; then
    error "directory parameter required" || return $failure
  else
    _tmp_dir="$1"
  fi

  mkdir -p $_tmp_dir; result=$?
  trap "rm -rf $_tmp_dir" EXIT SIGINT SIGTERM

  if [ $result -ne $success ]; then
    error "Failed to create tmp directory" || return $failure
  fi
}
