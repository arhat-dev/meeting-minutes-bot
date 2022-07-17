{{-
/*
    Page template to render messsages as plain text
*/ -}}
{{- define "gen.new" -}}{{- end -}}

{{- define "gen.body" -}}
  {{- range .Messages -}}
    {{-  range .Spans -}}
      {{- .Text -}}
    {{- end -}}
    {{- nindent 0 "" -}}
  {{- end -}}
{{- end -}}
