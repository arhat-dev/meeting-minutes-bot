# Publisher `http`

Issue http requests using generated content

## Input

This publisher requires a yaml config as the input:

```yaml
params:
- name: custom name from message
  value: custom value from message

# request body
body: |
  custom request body formatted from messages
```

## Config

```yaml
# target url the request will hit, you can use template here
# reference values from spec parsed from the yaml input
url: |-
  https://api.example.com?
  {{- range .Params -}}
    {{- .Name -}}={{- .Value -}}&
  {{- end -}}

# http method, one of [GET, HEAD, POST, OPTIONS, PUT, PATCH, DELETE]
method: |-
  {{- if gt (len .Body) 0 -}}
  POST
  {{- else -}}
  GET
  {{- end -}}

# headers for the request
headers:
- name: You-Can-Also-Use-Template-Here
  # reference values from spec parsed from the yaml input
  value: |-
    {{- .Params[0].value -}}

# client tls settings
tls: {}

# render response using go template
responseTemplate: |
  {{- jqBytes ".[0]" .Body -}}
```

## Reference Commands Mapping

```yaml
/discuss:
  as: /prepare
  description: prepare a http request
/end:
  as: /request
  description: Issue request with prepared arguments
# keep `/cancel` command as is
# /cancel: {}

# disable following commands since they are not used

/edit: {}
/list: {}
/delete: {}
/start: {}
/continue: {}
/ignore: {}
/include: {}
```
