# mbot

[![CI](https://github.com/arhat-dev/mbot/workflows/CI/badge.svg)](https://github.com/arhat-dev/mbot/actions?query=workflow%3ACI)
[![Build](https://github.com/arhat-dev/mbot/workflows/Build/badge.svg)](https://github.com/arhat-dev/mbot/actions?query=workflow%3ABuild)
[![PkgGoDev](https://pkg.go.dev/badge/arhat.dev/mbot)](https://pkg.go.dev/arhat.dev/mbot)
[![GoReportCard](https://goreportcard.com/badge/arhat.dev/mbot)](https://goreportcard.com/report/arhat.dev/mbot)
[![codecov](https://codecov.io/gh/arhat-dev/mbot/branch/master/graph/badge.svg)](https://codecov.io/gh/arhat-dev/mbot)

Build your chat bot declaratively.

## What?

`mbot` (previously `meeting-minutes-bot`) was a bot built for recording discussions on specifiec topics, after the initial implementation and usage, we came to aware almost all existing bots can be derived from this bot due to its ability to record meeting minutes.

## Features

- [x] Automatic post generation from messages
- [x] Automatic file uploading for content sharing
- [ ] Automatic web archiving for links in message
- [x] Basic post management (list, delete, edit)

## Support Matrix

- Chat Platform
  - [x] `telegram`
- [Storage](./docs/storage/README.md)
  - [x] [`s3`](./docs/storage/s3.md) (no presign support, requires public read access)
  - [x] [`telegraph`](./docs/storage/telegraph.md)
- [Generator](./docs/generator/README.md)
  - [x] [`gotemplate`](./docs/generator/gotemplate.md)
  - [x] [`file`](./docs/generator/file.md)
- [Publisher](./docs/publisher/README.md)
  - [x] [`telegraph`](./docs/publisher/telegraph.md)
  - [x] [`file`](./docs/publisher/file.md)
  - [x] [`interpreter`](./docs/publisher/interpreter.md)
  - [x] [`http`](./docs/publisher/http.md)
- Web Archiver
  - [ ] `cdp` (chrome dev tools protocol with headless chromium)

## Bot Commands

- `/discuss TOPIC`
  - Request starting a new session around `TOPIC`, the `TOPIC` will be the title of the published post
    - Once you finished the guided operation, you will get a `POST_KEY`
- `/continue POST_KEY`
  - Request continuing previous session, use corresponding `POST_KEY` for your `TOPIC`
- `/ignore` (During Session)
  - Use this command as a reply to the message you want to omit in the final post.
- `/include` (During Session)
  - Use this command as a reply to a message you want to include in the final post but sent before the session start.
- `/edit` (Private Message Only)
  - Edit your posts with external application
- `/list` (Private Message Only)
  - Request listing your posts
- `/delete url1{ url2 ... urlN}` (Private Message Only)
  - Request deleting post(s) using their `POST_KEY`s, you can provide multiple space separated URLs
- `/end`
  - End current session
- `/cancel`
  - Cancel pending request

__NOTE:__ You have to manage `POST_KEY`s for your `TOPIC`s

## State Machine

```mermaid
<!-- TODO -->
```

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
