{{- define "message.author.link" -}}
  {{- if .AuthorLink -}}
    {{- if .MessageLink -}}{{- indent 1 "" -}}{{- end -}}
    <a href="{{- .AuthorLink -}}">
  {{- end -}}
  {{- .Author -}}
  {{- if .AuthorLink -}}</a>{{- end -}}

  {{- if .ChatName -}}
    {{- indent 1 "" -}}@{{- indent 1 "" -}}
    {{- if .ChatLink -}}
      <a href="{{- .ChatLink -}}">
    {{- end -}}

      {{- .ChatName -}}

    {{- if .ChatLink -}}
      </a>
    {{- end -}}
  {{- end -}}

  {{- if and .IsForwarded (gt (len .OriginalAuthor) 0) -}}
    {{- indent 1 "" -}}From{{- indent 1 "" -}}
    {{- if .OriginalAuthorLink -}}
      <a href="{{- .OriginalAuthorLink -}}">
    {{- end -}}
      {{- .OriginalAuthor -}}
    {{- if .OriginalAuthorLink -}}
      </a>
    {{- end -}}

    {{- if .OriginalChatName -}}
      @{{- indent 1 "" -}}
      {{- if .OriginalChatLink -}}
        <a href="{{- .OriginalChatLink -}}">
      {{- end -}}
        {{- .OriginalChatName -}}
      {{- if .OriginalChatLink -}}
        </a>
      {{- end -}}
    {{- end -}}
  {{- end -}}

{{- end -}}

{{- define "message.header" -}}
<p>
  {{- /* Message Link */ -}}

  {{- if .MessageLink -}}
    <a href="{{- .MessageLink -}}">[Message]</a>
  {{- end -}}

  {{- /* Message Author Info */ -}}
  {{- template "message.author.link" . -}}
</p>
{{- end -}} {{- /* define */ -}}
