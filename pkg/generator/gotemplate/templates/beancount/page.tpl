{{-
/*
    Page template to render beancount entries
*/ -}}
{{- define "page.header" -}}{{- end -}}

{{- define "page.body" -}}
  {{- range .Messages -}}
    {{- template "message.body" . -}}
    {{- nindent 0 "" -}}
  {{- end -}}
{{- end -}}
