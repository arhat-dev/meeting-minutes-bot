# mbot

[![CI](https://github.com/arhat-dev/mbot/workflows/CI/badge.svg)](https://github.com/arhat-dev/mbot/actions?query=workflow%3ACI)
[![Build](https://github.com/arhat-dev/mbot/workflows/Build/badge.svg)](https://github.com/arhat-dev/mbot/actions?query=workflow%3ABuild)
[![PkgGoDev](https://pkg.go.dev/badge/arhat.dev/mbot)](https://pkg.go.dev/arhat.dev/mbot)
[![GoReportCard](https://goreportcard.com/badge/arhat.dev/mbot)](https://goreportcard.com/report/arhat.dev/mbot)
[![codecov](https://codecov.io/gh/arhat-dev/mbot/branch/master/graph/badge.svg)](https://codecov.io/gh/arhat-dev/mbot)

Build your chat bot with mbot

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

## Build

```bash
make mbot
```

You can find `mbot` in `./build`

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
    # driver name, one of [s3, telegraph]
  - telegraph:
      # max data size for uploading, size limit in bytes
      #
      # defaults to 0, no size limit
      maxUploadSize: 5242880 # 5MB
      # mime type match (regex), only matched data will use this storage driver
      #
      # defaults to "", match all
      mimeMatch: image/.*
      # read ./docs/storage/{DRIVER}.md to find out config options
      config: {}
  - s3:
      maxUploadSize: 0 # no limit
      mimeMatch: "" # match all
      config: {}

  # currently not supported
  # webarchiver:
  #   cdp: {}

  # post generator
  # read ./docs/generator/{DRIVER}.md to find out config options
  generator:
    # one of [gotemplate, file]
    gotemplate: {}

  # post publisher
  # read ./docs/publisher/{DRIVER}.md to find out config options
  publisher:
    # one of [telegraph, file, interpreter, http]
    interpreter: {}

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
    botToken@env: ${MY_TELEGRAM_BOT_TOKEN}

    # currently not tested, feel free to file issues if you find it not working
    webhook:
      enabled: false
      # the final webhook url is `$.app.publicBaseURL` + `$.bots.telegram.webhook.path`
      path: /telegram
      maxConnections: 100
      # required to provide tls public key when using self-signed certificate
      # tlsPublicKey@file: path/to/tls.pub
```

__NOTE:__ You can reference environment variables in config file (e.g. `${FOO}`, `$BAR`)

## Run

```bash
/path/to/built/mbot -c /path/to/your/config.yaml
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
