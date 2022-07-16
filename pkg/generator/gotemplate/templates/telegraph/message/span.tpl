{{- define "message.span" -}}
  {{- if .IsMedia -}}
    {{- template "media" . -}}
  {{- else -}}
    {{- template "text" . -}}
  {{- end -}}
{{- end -}}

{{- define "message.span.replied" -}}
  {{- if .IsMedia -}}
    {{- template "media" . -}}
  {{- else -}}
    {{- template "text.replied" . -}}
  {{- end -}}
{{- end -}}
