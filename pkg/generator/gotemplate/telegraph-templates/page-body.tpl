{{-
/*
    Page template to render messsages
    Go to https://telegra.ph/api and find supported tags
*/ -}}

{{- define "page.body" -}}

{{- if .Messages -}}
  {{- range .Messages -}}
    {{- if not .IsPrivateMessage -}}
      {{- template "message.header" . -}}
    {{- end -}}
    {{- template "message.body" . -}}
    {{- template "message.footer" . -}}
  {{- end -}} {{- /* Message */ -}}
{{- end -}} {{- /* Page */ -}}

{{- template "page.footer" . -}}

{{- end -}}
