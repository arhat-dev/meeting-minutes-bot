{{- define "text.replied" -}}
  
  {{- /* 
        we use blockquote to format reply message, inner blockquote will break format
        so we just ignore the blockquote style here

        NOTE: telegraph doesn't support multiple styles on the same range of text
  */ -}}

  {{- if .IsBold -}}
    <strong>{{- template "text.value" . -}}</strong>
  {{- else if .IsItalic -}}
    <em>{{- template "text.value" . -}}</em>
  {{- else if .IsStrikethrough -}}
    <del>{{- template "text.value" . -}}</del>
  {{- else if .IsUnderline -}}
    <u>{{- template "text.value" . -}}</u>
  {{- else if .IsCode -}}
    <code>{{- template "text.value" . -}}</code>
  {{- else if .IsPre -}}
    <pre>{{- template "text.value" . -}}</pre>
  {{- else -}}
    {{- template "text.value" . -}}
  {{- end -}}

{{- end -}}

{{- define "text" -}}
  {{- if .IsBlockquote -}}
    <blockquote>{{- template "text.value" . -}}</blockquote>
  {{- else -}}
    {{- template "text.replied" . -}}
  {{- end -}}
{{- end -}}

{{- define "text.value" -}}

  {{- if .IsLink -}}

    {{- if .IsEmail -}}
      <a href="mailto:{{- .Text -}}">{{- .Text -}}</a>
    {{- else if .IsPhoneNumber -}}
      <a href="tel:{{- .Text -}}">{{- .Text -}}</a>
    {{- else if .IsMention -}}
      <a href="{{- .URL -}}">{{- .Text -}}</a>
    {{- else if .IsURL -}}
      <a href="{{- .URL -}}">{{- .Text -}}</a>

      {{- if .WebArchiveURL -}}
      {{- indent 1 "" -}}
      <a href="{{- .WebArchiveURL -}}">[archive]</a>
      {{- end -}}

      {{- if .WebArchiveScreenshotURL -}}
      {{- indent 1 "" -}}
      <a href="{{- .WebArchiveScreenshotURL -}}">[screenshot]</a>
      {{- end -}}

    {{- end -}}

  {{- else -}}
    {{- .Text -}}
  {{- end -}}

{{- end -}}