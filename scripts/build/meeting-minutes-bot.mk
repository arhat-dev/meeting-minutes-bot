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

# native
meeting-minutes-bot:
	sh scripts/build/build.sh $@

# linux
meeting-minutes-bot.linux.x86:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.amd64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.armv5:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.armv6:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.armv7:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.arm64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.mips:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.mipshf:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.mipsle:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.mipslehf:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.mips64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.mips64hf:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.mips64le:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.mips64lehf:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.ppc64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.ppc64le:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.s390x:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.riscv64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.linux.all: \
	meeting-minutes-bot.linux.x86 \
	meeting-minutes-bot.linux.amd64 \
	meeting-minutes-bot.linux.armv5 \
	meeting-minutes-bot.linux.armv6 \
	meeting-minutes-bot.linux.armv7 \
	meeting-minutes-bot.linux.arm64 \
	meeting-minutes-bot.linux.mips \
	meeting-minutes-bot.linux.mipshf \
	meeting-minutes-bot.linux.mipsle \
	meeting-minutes-bot.linux.mipslehf \
	meeting-minutes-bot.linux.mips64 \
	meeting-minutes-bot.linux.mips64hf \
	meeting-minutes-bot.linux.mips64le \
	meeting-minutes-bot.linux.mips64lehf \
	meeting-minutes-bot.linux.ppc64 \
	meeting-minutes-bot.linux.ppc64le \
	meeting-minutes-bot.linux.s390x \
	meeting-minutes-bot.linux.riscv64

meeting-minutes-bot.darwin.amd64:
	sh scripts/build/build.sh $@

# # currently darwin/arm64 build will fail due to golang link error
# meeting-minutes-bot.darwin.arm64:
# 	sh scripts/build/build.sh $@

meeting-minutes-bot.darwin.all: \
	meeting-minutes-bot.darwin.amd64

meeting-minutes-bot.windows.x86:
	sh scripts/build/build.sh $@

meeting-minutes-bot.windows.amd64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.windows.armv5:
	sh scripts/build/build.sh $@

meeting-minutes-bot.windows.armv6:
	sh scripts/build/build.sh $@

meeting-minutes-bot.windows.armv7:
	sh scripts/build/build.sh $@

# # currently no support for windows/arm64
# meeting-minutes-bot.windows.arm64:
# 	sh scripts/build/build.sh $@

meeting-minutes-bot.windows.all: \
	meeting-minutes-bot.windows.x86 \
	meeting-minutes-bot.windows.amd64 \
	meeting-minutes-bot.windows.armv5 \
	meeting-minutes-bot.windows.armv6 \
	meeting-minutes-bot.windows.armv7

# # android build requires android sdk
# meeting-minutes-bot.android.amd64:
# 	sh scripts/build/build.sh $@

# meeting-minutes-bot.android.x86:
# 	sh scripts/build/build.sh $@

# meeting-minutes-bot.android.armv5:
# 	sh scripts/build/build.sh $@

# meeting-minutes-bot.android.armv6:
# 	sh scripts/build/build.sh $@

# meeting-minutes-bot.android.armv7:
# 	sh scripts/build/build.sh $@

# meeting-minutes-bot.android.arm64:
# 	sh scripts/build/build.sh $@

# meeting-minutes-bot.android.all: \
# 	meeting-minutes-bot.android.amd64 \
# 	meeting-minutes-bot.android.arm64 \
# 	meeting-minutes-bot.android.x86 \
# 	meeting-minutes-bot.android.armv7 \
# 	meeting-minutes-bot.android.armv5 \
# 	meeting-minutes-bot.android.armv6

meeting-minutes-bot.freebsd.amd64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.freebsd.x86:
	sh scripts/build/build.sh $@

meeting-minutes-bot.freebsd.armv5:
	sh scripts/build/build.sh $@

meeting-minutes-bot.freebsd.armv6:
	sh scripts/build/build.sh $@

meeting-minutes-bot.freebsd.armv7:
	sh scripts/build/build.sh $@

meeting-minutes-bot.freebsd.arm64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.freebsd.all: \
	meeting-minutes-bot.freebsd.amd64 \
	meeting-minutes-bot.freebsd.arm64 \
	meeting-minutes-bot.freebsd.armv7 \
	meeting-minutes-bot.freebsd.x86 \
	meeting-minutes-bot.freebsd.armv5 \
	meeting-minutes-bot.freebsd.armv6

meeting-minutes-bot.netbsd.amd64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.netbsd.x86:
	sh scripts/build/build.sh $@

meeting-minutes-bot.netbsd.armv5:
	sh scripts/build/build.sh $@

meeting-minutes-bot.netbsd.armv6:
	sh scripts/build/build.sh $@

meeting-minutes-bot.netbsd.armv7:
	sh scripts/build/build.sh $@

meeting-minutes-bot.netbsd.arm64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.netbsd.all: \
	meeting-minutes-bot.netbsd.amd64 \
	meeting-minutes-bot.netbsd.arm64 \
	meeting-minutes-bot.netbsd.armv7 \
	meeting-minutes-bot.netbsd.x86 \
	meeting-minutes-bot.netbsd.armv5 \
	meeting-minutes-bot.netbsd.armv6

meeting-minutes-bot.openbsd.amd64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.openbsd.x86:
	sh scripts/build/build.sh $@

meeting-minutes-bot.openbsd.armv5:
	sh scripts/build/build.sh $@

meeting-minutes-bot.openbsd.armv6:
	sh scripts/build/build.sh $@

meeting-minutes-bot.openbsd.armv7:
	sh scripts/build/build.sh $@

meeting-minutes-bot.openbsd.arm64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.openbsd.all: \
	meeting-minutes-bot.openbsd.amd64 \
	meeting-minutes-bot.openbsd.arm64 \
	meeting-minutes-bot.openbsd.armv7 \
	meeting-minutes-bot.openbsd.x86 \
	meeting-minutes-bot.openbsd.armv5 \
	meeting-minutes-bot.openbsd.armv6

meeting-minutes-bot.solaris.amd64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.aix.ppc64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.dragonfly.amd64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.plan9.amd64:
	sh scripts/build/build.sh $@

meeting-minutes-bot.plan9.x86:
	sh scripts/build/build.sh $@

meeting-minutes-bot.plan9.armv5:
	sh scripts/build/build.sh $@

meeting-minutes-bot.plan9.armv6:
	sh scripts/build/build.sh $@

meeting-minutes-bot.plan9.armv7:
	sh scripts/build/build.sh $@

meeting-minutes-bot.plan9.all: \
	meeting-minutes-bot.plan9.amd64 \
	meeting-minutes-bot.plan9.armv7 \
	meeting-minutes-bot.plan9.x86 \
	meeting-minutes-bot.plan9.armv5 \
	meeting-minutes-bot.plan9.armv6
