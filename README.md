# Meeting Minutes Bot

[![CI](https://github.com/arhat-dev/meeting-minutes-bot/workflows/CI/badge.svg)](https://github.com/arhat-dev/meeting-minutes-bot/actions?query=workflow%3ACI)
[![Build](https://github.com/arhat-dev/meeting-minutes-bot/workflows/Build/badge.svg)](https://github.com/arhat-dev/meeting-minutes-bot/actions?query=workflow%3ABuild)
[![PkgGoDev](https://pkg.go.dev/badge/arhat.dev/meeting-minutes-bot)](https://pkg.go.dev/arhat.dev/meeting-minutes-bot)
[![GoReportCard](https://goreportcard.com/badge/arhat.dev/meeting-minutes-bot)](https://goreportcard.com/report/arhat.dev/meeting-minutes-bot)
[![codecov](https://codecov.io/gh/arhat-dev/meeting-minutes-bot/branch/master/graph/badge.svg)](https://codecov.io/gh/arhat-dev/meeting-minutes-bot)

Build your KnowledgeBase in Chat

## Features

- [x] Automatic post generation from messages
- [ ] Automatic post update on message update
- [x] Automatic file uploading for content sharing
- [ ] Automatic web archiving for links in message
- [x] Basic post management (list, delete, edit)

## Support Matrix

- Chat Platform
  - [x] `telegram`
- Storage
  - [x] `s3` (no presign support, requires public read access)
- Post Generator
  - [x] [`gotemplate`](./docs/generator/gotemplate.md)
- Post Publisher
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

## Build

```bash
make meeting-minutes-bot
```

You can find `meeting-minutes-bot` in `./build`

## Config

```yaml
app:
  log:
  - level: verbose
    file: stderr

  # public facing base url for bots webhooks
  publicBaseURL: https://bot.example.com/base/path

  # tcp listen address for webhooks
  listen: :8080

  # tls settings for webhook (not tested)
  tls:
    enabled: false

  storage:
    # currently only supports `s3`, leave it empty to disable automatic file uploading
    driver: s3
    # read ./docs/storage/{DRIVER}.md to find out config options
    config: {}

  # currently not supported
  webarchiver:
    driver: ""

  # post generator
  generator:
    # currently only supports `gotemplate`
    driver: gotemplate
    # read ./docs/generator/{DRIVER}.md to find out config options
    config: {}

  # post publisher
  publisher:
    # one of [telegraph, file, interpreter, http]
    driver: interpreter
    # read ./docs/publisher/{DRIVER}.md to find out config options
    config: {}

bots:
  # rename default commands
  globalCommandsMapping:
    /discuss:
      as: /prepare
      description: prepare script for interpreter
    /end:
      as: /run
      description: run the prepared script
    # keep `/cancel` command as is
    #/cancel: {}

    # disable commands with emtpy body (DO NOT use null)

    /edit: {}
    /list: {}
    /delete: {}
    /start: {}
    /continue: {}
    /ignore: {}
    /include: {}

  telegram:
    # telegram api endpoint, NOT a URL
    endpoint: api.telegram.org
    # the bot token provided by the @BotFather
    botToken: ${MY_TELEGRAM_BOT_TOKEN}

    # currently not tested, feel free to file issues if you find it not working
    webhook:
      enabled: false
      # the final webhook url is `$.app.publicBaseURL` + `$.bots.telegram.webhook.path`
      path: /telegram
      maxConnections: 100
      # required to provide tls public key when using self-signed certificate
      # tlsPublicKey:
      # tlsPublicKeyData:
```

__NOTE:__ You can reference environment variables in config file (e.g. `${FOO}`, `$BAR`)

## Run

```bash
/path/to/built/meeting-minutes-bot -c /path/to/your/config.yaml
```

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
