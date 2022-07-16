{{-
/*
    Message template to render single message as a beancount entry
    https://beancount.github.io/docs/beancount_language_syntax.html
*/ -}}

{{- define "beancount.entry" -}}
{{- /* TODO */ -}}
{{- fail "unimplemented" -}}
{{- end -}}

{{- define "message.body" -}}
  {{- .Timestamp.Format "2006-01-02" -}}
  {{- template "beancount.entry" .Spans | indent 1 -}}
{{- end -}}
