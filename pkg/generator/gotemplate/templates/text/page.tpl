{{-
/*
    Page template to render messsages as plain text
*/ -}}
{{- define "page.header" -}}{{- end -}}

{{- define "page.body" -}}
  {{- range .Messages -}}
    {{-  range .Entities -}}
      {{- .Text -}}
    {{- end -}}
    {{- nindent 0 "" -}}
  {{- end -}}
{{- end -}}
