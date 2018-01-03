#!/bin/bash
#
# Copyright (c) 2017 by Cisco Systems, Inc.
# All rights reserved.
#
declare -ri success=0
declare -ri failure=1

declare ID="centos"
declare VERSION_ID="6"

if [[ -r /etc/os-release ]]; then
  source /etc/os-release
elif [[ -r /etc/system-release-cpe ]]; then
  ID="$(awk -F: '{print $3}' /etc/system-release-cpe)"
  VERSION_ID="$(awk -F: '{print $5}' /etc/system-release-cpe)"
fi

function die() {
  printf "Error: %s\n" "$@"
  exit $failure
}

function user_exists() {
  egrep -q "^${1}:" /etc/passwd
}

function group_exists() {
  egrep -q "^${1}:" /etc/group
}

set -x
