# Generator `gotemplate`

Render messages with [go template](https://golang.org/pkg/text/template/)

## Available template functions

- [Built-in functions](https://golang.org/pkg/text/template/#hdr-Functions)
- [Sprig functions](https://masterminds.github.io/sprig/)
- `jq <query> <string-data>`
- `jqBytes <query> <[]byte-data>`

## Requirements

- A valid template __MUST__ define `gen.new`, `gen.continue` and `gen.body`
  - The `gen.new` template is executed when calling `generator.New()`
    - It will be executed only once for each post
  - The `gen.body` template will be executed when calling `generator.Append()`
    - It will be executed multiple times if you `/resume` the session

## Config

```yaml
# available built-in templates are [telegraph, text, beancount, http-req-spec]
builtinTemplate: telegraph

# custom template directory, all files in this directory will be treated as the template
# leave it empty to use built-in templates
templatesDir: /path/to/templates/dir
# `html` or `text`
outputFormat: html
```
