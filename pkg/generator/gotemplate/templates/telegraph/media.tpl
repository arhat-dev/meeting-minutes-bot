{{- define "media" -}}

{{- if .IsImage -}}
  {{- template "image" . -}}
{{- else if .IsVideo -}}
  {{- template "video" . -}}
{{- else if .IsAudio -}}
  {{- template "video" . -}}
{{- else if .IsVoice -}}
  {{- template "video" . -}}
{{- else if .IsFile -}}
  <a href="{{- .URL -}}">[File]
    {{- if .Filename -}}
      {{- .Filename | indent 1 -}}
    {{- end -}}
  </a>

  {{- if .Caption -}}
    {{- nindent 0 "" -}}
    {{- range .Caption -}}
      {{- template "message.span" . -}}
    {{- end -}}
  {{- end -}}

{{- end -}}

{{- end -}} {{- /* define */ -}}


{{- define "video" -}}
<figure>
{{- /* no newline comment */ -}}
  <video src="{{- .URL -}}"></video>
  {{- if .Caption -}}
    <figcaption>
      {{- range .Caption -}}
        {{- template "message.span" . -}}
      {{- end -}}
    </figcaption>
  {{- end -}}
</figure>
{{- end -}}

{{- define "image" -}}
<figure>
{{- /* no newline comment */ -}}
  <img src="{{- .URL -}}"></img>
  {{- if .Caption -}}
    <figcaption>
      {{- range .Caption -}}
        {{- template "message.span" . -}}
      {{- end -}}
    </figcaption>
  {{- end -}}
</figure>
{{- end -}}