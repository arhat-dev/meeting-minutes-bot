#!/bin/sh

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

set -ex

_download_go_pakage() {
  GO111MODULE=off go get -u -v -d "$1"
}

_install_go_bin() {
  package="$1"
  cmd_dir="$2"
  bin="$3"

  # download
  temp_dir="$(mktemp -d)"
  cd "${temp_dir}"
  GO111MODULE=on go get -d -u "${package}"
  cd -
  rmdir "${temp_dir}"

  # build
  cd "${GOPATH}/pkg/mod/${package}"
  # TODO: currently the go.sum in github.com/deepmap/oapi-codegen is not synced
  #       once fixed in that repo, remove -mod=mod
  GO111MODULE=on go build -mod=mod -o "${bin}" "${cmd_dir}"
  cd -
}

install_tools_go() {
  GOPATH=$(go env GOPATH)
  export GOPATH

  GOOS=$(go env GOHOSTOS)
  GOARCH=$(go env GOHOSTARCH)
  export GOOS
  export GOARCH

  cd "$(mktemp -d)"
  _download_go_pakage github.com/gogo/protobuf/proto
  _download_go_pakage github.com/gogo/protobuf/gogoproto

  _install_go_bin "github.com/deepmap/oapi-codegen@v1.6.1" "./cmd/oapi-codegen" "${GOPATH}/bin/oapi-codegen"
  cd -
}

install_tools_go
