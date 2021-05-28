
{{-
/*
    Render telegram message in html format
    https://core.telegram.org/bots/api/#formatting-options

*/ -}}

{{- define "message" -}}
{{- range . -}}
  {{- if .IsText -}}
    {{- .Text -}}
  {{- else if .IsBold -}}
    <strong>{{- .Text -}}</strong>
  {{- else if .IsItalic -}}
    <em>{{- .Text -}}</em>
  {{- else if .IsStrikethrough -}}
    <del>{{- .Text -}}</del>
  {{- else if .IsUnderline -}}
    <u>{{- .Text -}}</u>
  {{- else if .IsPre -}}
    <pre>{{- .Text -}}</pre>
  {{- else if .IsCode -}}
    <code>{{- .Text -}}</code>
  {{- else if .IsEmail -}}
    <a href="mailto:{{- .Text -}}">{{- .Text -}}</a>
  {{- else if .IsPhoneNumber -}}
    <a href="tel:{{- .Text -}}">{{- .Text -}}</a>
  {{- else if .IsURL -}}
    <a href="{{- index .Params "url" -}}">{{- .Text -}}</a>
  {{- end -}}
{{- end -}}
{{- end -}}
