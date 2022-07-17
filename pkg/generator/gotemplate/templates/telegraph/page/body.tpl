{{-
/*
    Page body template to render messsages
    Go to https://telegra.ph/api and find supported tags
*/ -}}

{{- define "gen.body" -}}

{{- range $_, $msg := .Messages -}}

  {{- if not .IsPrivateMessage -}}
    {{- template "message.header" $msg -}}
  {{- end -}}

  {{- template "message.body" $msg -}}
  {{- template "message.footer" $msg -}}

{{- end -}} {{- /* Messages */ -}}

{{- template "page.footer" . -}}

{{- end -}}
