# mbot

[![CICD](https://github.com/arhat-dev/mbot/workflows/CI/badge.svg)](https://github.com/arhat-dev/mbot/actions?query=workflow%3ACI)
[![PkgGoDev](https://pkg.go.dev/badge/arhat.dev/mbot)](https://pkg.go.dev/arhat.dev/mbot)
[![GoReportCard](https://goreportcard.com/badge/arhat.dev/mbot)](https://goreportcard.com/report/arhat.dev/mbot)
[![Coverage](https://badge.arhat.dev/sonar/coverage/arhat-dev_mbot?branch=master&token=5ef94767c23a62c0c9ee2584887506d7)](https://sonar.arhat.dev/dashboard?id=arhat-dev_mbot)

Build your chat bot declaratively.

## What?

`mbot` (previously `meeting-minutes-bot`) was a bot built for recording discussions on specifiec topics, after the initial implementation and usage, we came to aware almost all existing bots can be derived from this bot due to its ability to record meeting minutes.

## Concepts

- [`bot`](./docs/bot/README.md)
- [`generator`](./docs/generator/README.md)
- [`storage`](./docs/storage/README.md)
- [`publisher`](./docs/publisher/README.md)

## Support Matrix

- Chat Platforms
  - [ ] `discord`
  - [ ] `github`
  - [ ] `gitlab`
  - [ ] `gitter`
  - [ ] `irc`
  - [ ] `line`
  - [ ] `matrix`
  - [ ] `mattermost`
  - [ ] `reddit`
  - [ ] `slack`
  - [x] `telegram`
  - [ ] `vk`
  - [ ] `whatsapp`

- [Storage backends](./docs/storage/README.md)
  - [x] `router`
  - [x] `s3` (no presign support, requires public read access)
  - [x] `telegraph` (image and audio/video)

- [Generators](./docs/generator/README.md)
  - [ ] `archiver`
  - [ ] `chain`
  - [ ] `cron`
  - [ ] `exec`
  - [ ] `filter`
  - [x] `gotemplate`
  - [ ] `js`
  - [ ] `lua`
  - [x] `multigen`
  - [ ] `tengo`

- [Publishers](./docs/publisher/README.md)
  - [x] `authorized`
  - [ ] `bot`
  - [x] `file`
  - [x] `http`
  - [ ] `mq`
  - [ ] `multipub`
  - [x] `telegraph`

## Run

```bash
/path/to/mbot -c /path/to/config.yaml
```

see [cicd/test/config.yml](./cicd/test/config.yml) for config example

## LICENSE

```text
Copyright 2021 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
