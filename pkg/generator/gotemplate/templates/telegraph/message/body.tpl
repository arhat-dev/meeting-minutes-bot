{{- define "message.body" -}}

<p>
{{- range $_, $span := .Spans -}}
  {{- template "message.span" $span -}}
{{- end -}}

{{- if .IsReply -}}
  {{- $replyTo := findMessage .ReplyTo -}}
  {{- if $replyTo -}}
    {{- nindent 0 "" -}}
    (in reply to
      {{- indent 1 "" -}}
      {{- template "message.author.link" $replyTo -}}
      {{- if $replyTo.MessageLink -}}
        {{- indent 1 "" -}}
        <a href="{{- $replyTo.MessageLink -}}">[Message]</a>
      {{- end -}}
    )
    {{- /* nospace comment */ -}}
    <blockquote>
      {{- template "message.body.replied" $replyTo -}}
    </blockquote>
  {{- end -}}
{{- end -}}
</p>

{{- end -}} {{- /* define */ -}}

{{- define "message.body.replied" -}}

<p>
{{- range $_, $span := .Spans -}}
  {{- template "message.span.replied" $span -}}
{{- end -}}
</p>

{{- end -}} {{- /* define */ -}}

