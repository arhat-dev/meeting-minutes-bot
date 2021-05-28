# Meeting Minutes Bot

[![CI](https://github.com/arhat-dev/meeting-minutes-bot/workflows/CI/badge.svg)](https://github.com/arhat-dev/meeting-minutes-bot/actions?query=workflow%3ACI)
[![Build](https://github.com/arhat-dev/meeting-minutes-bot/workflows/Build/badge.svg)](https://github.com/arhat-dev/meeting-minutes-bot/actions?query=workflow%3ABuild)
[![PkgGoDev](https://pkg.go.dev/badge/arhat.dev/meeting-minutes-bot)](https://pkg.go.dev/arhat.dev/meeting-minutes-bot)
[![GoReportCard](https://goreportcard.com/badge/arhat.dev/meeting-minutes-bot)](https://goreportcard.com/report/arhat.dev/meeting-minutes-bot)
[![codecov](https://codecov.io/gh/arhat-dev/meeting-minutes-bot/branch/master/graph/badge.svg)](https://codecov.io/gh/arhat-dev/meeting-minutes-bot)

Build your knowledgebase in Chat

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
- Web Archiver
  - [ ] `cdp` (chrome dev tools protocol with headless chromium)

## Bot Commands

- [x] `/discuss TOPIC`
  - Start a new session around `TOPIC`, the `TOPIC` will be the title of the published post
  - Once you finished the guided operation, you will get a `POST_URL`
- [x] `/continue POST_URL`
  - Continue previous session, use corresponding `POST_URL` for your `TOPIC`
- [x] `/ignore` (During Session)
  - Use this command as a reply to the message you want to omit in the final post.
- [x] `/include` (During Session)
  - Use this command as a reply to a message you want to include in the final post but sent before the session start.
- [x] `/edit` (Private Message Only)
  - Edit your posts with external application
- [x] `/list` (Private Message Only)
  - List your posts
- [x] `/delete url1{ url2 ... urlN}` (Private Message Only)
  - Delete post(s) using their `POST_URL`s, you can provide multiple space separated URLs
- [x] `/end` - end current session or cancel current operation

__NOTE:__ You have to manage `POST_URL`s for your `TOPIC`s

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
    config:
      # s3 service endpoint, a URL, scheme MUST be `http` or `https`
      endpointURL: https://s3.example.com
      # bucket region
      region: us-east-1
      # bucket name
      bucket: example
      # path in
      basePath: foo/bar
      # access key (required)
      accessKeyID: ${MY_S3_ACCESS_KEY}
      # access key secret (required)
      accessKeySecret: ${MY_S3_SECRET_KEY}

  # currently not supported
  webarchiver:
    driver: ""

  # post generator
  generator:
    # currently only supports `gotemplate`
    driver: gotemplate
    config:
      # available built-in templates are [telegraph, text, beancount]
      builtinTemplate: telegraph
      # custom template directory, all files in this directory will be treated as template
      #templatesDir: /path/to/templates/dir
      # `html` or `text`
      #outputFormat: html

  # post publisher
  publisher:
    # one of [telegraph, file, interpreter]
    driver: telegraph
    # read ./docs/publisher/{DRIVER}.md to find out config options
    #config: {}

bots:
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
