# Generator `gotemplate`

Render messages with [go template](https://golang.org/pkg/text/template/)

## Available template functions

- [Built-in functions](https://golang.org/pkg/text/template/#hdr-Functions)
- [Sprig functions](https://masterminds.github.io/sprig/)
- `jq <query> <string-data>`
- `jqBytes <query> <[]byte-data>`

## Requirements

- A valid template __MUST__ have `{{ define "page.header" }}` and `{{ define "page.body" }}`
  - The `page.header` template is executed when calling `generator.Publish()`
    - It will be executed only once for each post
  - The `page.body` template will be executed when calling `generator.Append()`
    - It will be executed multiple times if you `/continue` the session
