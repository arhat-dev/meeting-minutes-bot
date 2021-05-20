#!/bin/sh

# Copyright 2021 The arhat.dev Authors.
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

image="$1"
src="$2"
dest="$3"

ctr_id="$(docker create "${image}" :)"
docker cp "${ctr_id}:${src}" "${dest}"
docker rm -f "${ctr_id}"
