#!/bin/sh
# shellcheck disable=SC2039

# Copyright 2020 The arhat.dev Authors.
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

set -e

. scripts/version.sh
. scripts/build/common.sh

_install_deps() {
  echo "${INSTALL}"
  eval "${INSTALL}"
}

_build() {
  echo "$1"
  eval "$1"
}

meeting_minutes_bot() {
  # TODO: set mandatory tags and predefined tags for specific platforms
  _build "CGO_ENABLED=0 ${GOBUILD} -tags='nokube nocloud netgo ${PREDEFINED_BUILD_TAGS} ${TAGS}' ./cmd/meeting-minutes-bot"
}

COMP=$(printf "%s" "$@" | cut -d. -f1)
CMD=$(printf "%s" "$@" | tr '-' '_' | tr '.'  ' ')

# CMD format: {comp} {os} {arch}

GOOS="$(printf "%s" "$@" | cut -d. -f2 || true)"
ARCH="$(printf "%s" "$@" | cut -d. -f3 || true)"

if [ -z "${GOOS}" ] || [ "${GOOS}" = "$(printf "%s" "${COMP}")" ]; then
  # fallback to goos and goarch values
  GOOS="$(go env GOHOSTOS)"
  ARCH="$(go env GOHOSTARCH)"
fi

GOEXE=""
PREDEFINED_BUILD_TAGS=""
case "${GOOS}" in
  darwin)
    PREDEFINED_BUILD_TAGS=""
  ;;
  openbsd)
    PREDEFINED_BUILD_TAGS=""
  ;;
  netbsd)
    PREDEFINED_BUILD_TAGS=""
  ;;
  freebsd)
    PREDEFINED_BUILD_TAGS=""
  ;;
  plan9)
    PREDEFINED_BUILD_TAGS=""
  ;;
  aix)
    PREDEFINED_BUILD_TAGS=""
  ;;
  solaris)
    PREDEFINED_BUILD_TAGS=""
  ;;
  linux)
    case "${ARCH}" in
      mips64* | mipsle*)
        PREDEFINED_BUILD_TAGS=""
      ;;
      riscv64)
        PREDEFINED_BUILD_TAGS=""
      ;;
    esac
  ;;
  windows)
    GOEXE=".exe"
    case "${ARCH}" in
      arm*)
        PREDEFINED_BUILD_TAGS=""
      ;;
    esac
  ;;
esac

GO_LDFLAGS="-s -w \
  -X arhat.dev/meeting-minutes-bot/pkg/version.branch=${GIT_BRANCH} \
  -X arhat.dev/meeting-minutes-bot/pkg/version.commit=${GIT_COMMIT} \
  -X arhat.dev/meeting-minutes-bot/pkg/version.tag=${GIT_TAG} \
  -X arhat.dev/meeting-minutes-bot/pkg/version.arch=${ARCH} \
  -X arhat.dev/meeting-minutes-bot/pkg/version.goCompilerPlatform=$(go version | cut -d\  -f4)"

GOARM="$(_get_goarm "${ARCH}")"
if [ -z "${GOARM}" ]; then
  # this can happen if no ARCH specified
  GOARM="$(go env GOARM)"
fi

GOMIPS="$(_get_gomips "${ARCH}")"
if [ -z "${GOMIPS}" ]; then
  # this can happen if no ARCH specified
  GOMIPS="$(go env GOMIPS)"
fi

GOBUILD="GO111MODULE=on GOOS=${GOOS} GOARCH=$(_get_goarch "${ARCH}") \
  GOARM=${GOARM} GOMIPS=${GOMIPS} GOWASM=satconv,signext \
  ${CGO_FLAGS} \
  go build -trimpath -buildmode=${BUILD_MODE:-default} \
  -mod=vendor -ldflags='${GO_LDFLAGS}' -o build/$*${GOEXE}"

$CMD
