{{-
/*
    Page template to render beancount entries
*/ -}}
{{- define "gen.new" -}}{{- end -}}

{{- define "gen.body" -}}
  {{- range .Messages -}}
    {{- template "message.body" . -}}
    {{- nindent 0 "" -}}
  {{- end -}}
{{- end -}}
