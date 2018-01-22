#
# Copyright (c) 2018, Cisco Systems
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

set -e

export ARC_ROOT="$PWD/cli"
export debug=yes

function die() {
  local rc=$1; shift
  printf "Failed: $@, rc: $rc"
  exit $rc
}

function run() {
  local rc=0

  echo ""
  echo "$@: success expected"
  echo "-------------------------------------------------------"

  set +e
  "$@"; rc=$?
  set -e
  echo -e "\nresult: $rc"

  if [[ $rc -ne 0 ]]; then
    die $rc "$@"
  fi

  printf "\n\n"
}


function run_err() {
  local rc=0

  echo "$@: err expected"
  echo "-------------------------------------------------------"

  set +e
  "$@"; rc=$?
  set -e
  echo -e "\nresult: $rc"

  if [[ $rc -eq 0 ]]; then
    die $rc "$@"
  fi

  printf "\n\n"
}
