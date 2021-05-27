{{- define "message.replied.body" -}}

<p>
{{- range .Entities -}}
  {{- template "message.replied.entity" . -}}
{{- end -}}
</p>

{{- end -}} {{- /* define */ -}}


{{- define "message.body" -}}

<p>
{{- range .Entities -}}
  {{- template "message.entity" . -}}
{{- end -}}

{{- if .IsReply -}}
  {{- $msg_replied := findMessage .Messages .ReplyToMessageID -}}
  {{- if $msg_replied -}}
    {{- nindent 0 "" -}}
    {{- /* I will delete spaces :) */ -}}
    <blockquote>
    {{- /* I will delete spaces :) */ -}}
    (in reply to
      {{- indent 1 "" -}}
      {{- template "message.author.link" $msg_replied -}}
      {{- if $msg_replied.MessageURL -}}
        {{- indent 1 "" -}}
        <a href="{{- $msg_replied.MessageURL -}}">[Message]</a>
      {{- end -}}
    )
      {{- nindent 0 "" -}}
      {{- template "message.replied.body" $msg_replied -}}
    {{- /* I will delete spaces :) */ -}}
    </blockquote>
  {{- end -}}
{{- end -}}
</p>

{{- end -}} {{- /* define */ -}}
