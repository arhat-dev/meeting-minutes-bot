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

#
# linux
#
package.meeting-minutes-bot.deb.amd64:
	sh scripts/package/package.sh $@

package.meeting-minutes-bot.deb.armv6:
	sh scripts/package/package.sh $@

package.meeting-minutes-bot.deb.armv7:
	sh scripts/package/package.sh $@

package.meeting-minutes-bot.deb.arm64:
	sh scripts/package/package.sh $@

package.meeting-minutes-bot.deb.all: \
	package.meeting-minutes-bot.deb.amd64 \
	package.meeting-minutes-bot.deb.armv6 \
	package.meeting-minutes-bot.deb.armv7 \
	package.meeting-minutes-bot.deb.arm64

package.meeting-minutes-bot.rpm.amd64:
	sh scripts/package/package.sh $@

package.meeting-minutes-bot.rpm.armv7:
	sh scripts/package/package.sh $@

package.meeting-minutes-bot.rpm.arm64:
	sh scripts/package/package.sh $@

package.meeting-minutes-bot.rpm.all: \
	package.meeting-minutes-bot.rpm.amd64 \
	package.meeting-minutes-bot.rpm.armv7 \
	package.meeting-minutes-bot.rpm.arm64

package.meeting-minutes-bot.linux.all: \
	package.meeting-minutes-bot.deb.all \
	package.meeting-minutes-bot.rpm.all

#
# windows
#

package.meeting-minutes-bot.msi.amd64:
	sh scripts/package/package.sh $@

package.meeting-minutes-bot.msi.arm64:
	sh scripts/package/package.sh $@

package.meeting-minutes-bot.msi.all: \
	package.meeting-minutes-bot.msi.amd64 \
	package.meeting-minutes-bot.msi.arm64

package.meeting-minutes-bot.windows.all: \
	package.meeting-minutes-bot.msi.all

#
# darwin
#

package.meeting-minutes-bot.pkg.amd64:
	sh scripts/package/package.sh $@

package.meeting-minutes-bot.pkg.arm64:
	sh scripts/package/package.sh $@

package.meeting-minutes-bot.pkg.all: \
	package.meeting-minutes-bot.pkg.amd64 \
	package.meeting-minutes-bot.pkg.arm64

package.meeting-minutes-bot.darwin.all: \
	package.meeting-minutes-bot.pkg.all
