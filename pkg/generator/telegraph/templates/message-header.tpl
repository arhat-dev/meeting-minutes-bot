{{- define "message.author.link" -}}
  {{- if .AuthorURL -}}
    {{- if .MessageURL -}}{{- indent 1 "" -}}{{- end -}}
    <a href="{{- .AuthorURL -}}">
  {{- end -}}
  {{- .Author -}}
  {{- if .AuthorURL -}}</a>{{- end -}}

  {{- if .ChatName -}}
    {{- indent 1 "" -}}@{{- indent 1 "" -}}
    {{- if .ChatURL -}}
      <a href="{{- .ChatURL -}}">
    {{- end -}}

      {{- .ChatName -}}

    {{- if .ChatURL -}}
      </a>
    {{- end -}}
  {{- end -}}

  {{- if and .IsForwarded (gt (len .OriginalAuthor) 0) -}}
    {{- indent 1 "" -}}From{{- indent 1 "" -}}
    {{- if .OriginalAuthorURL -}}
      <a href="{{- .OriginalAuthorURL -}}">
    {{- end -}}
      {{- .OriginalAuthor -}}
    {{- if .OriginalAuthorURL -}}
      </a>
    {{- end -}}

    {{- if .OriginalChatName -}}
      @{{- indent 1 "" -}}
      {{- if .OriginalChatURL -}}
        <a href="{{- .OriginalChatURL -}}">
      {{- end -}}
        {{- .OriginalChatName -}}
      {{- if .OriginalChatURL -}}
        </a>
      {{- end -}}
    {{- end -}}
  {{- end -}}

{{- end -}}

{{- define "message.header" -}}
<p>
  {{- /* Message Link */ -}}

  {{- if .MessageURL -}}
    <a href="{{- .MessageURL -}}">[Link]</a>
  {{- end -}}

  {{- /* Message Author Info */ -}}
  {{- template "message.author.link" . -}}
</p>
{{- end -}} {{- /* define */ -}}
