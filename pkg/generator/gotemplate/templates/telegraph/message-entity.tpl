{{- define "message.entity" -}}
  {{- if .IsText -}}{{- .Text -}}{{- end -}}
  {{- if .IsBold -}}<strong>{{- .Text -}}</strong>{{- end -}}
  {{- if .IsItalic -}}<em>{{- .Text -}}</em>{{- end -}}
  {{- if .IsStrikethrough -}}<del>{{- .Text -}}</del>{{- end -}}
  {{- if .IsUnderline -}}<u>{{- .Text -}}</u>{{- end -}}
  {{- if .IsPre -}}<pre>{{- .Text -}}</pre>{{- end -}}
  {{- if .IsCode -}}<code>{{- .Text -}}</code>{{- end -}}
  {{- if .IsBlockquote -}}<blockquote>{{- .Text -}}</blockquote>{{- end -}}

  {{- if .IsEmail -}}<a href="mailto:{{- .Text -}}">{{- .Text -}}</a>{{- end -}}
  {{- if .IsPhoneNumber -}}<a href="tel:{{- .Text -}}">{{- .Text -}}</a>{{- end -}}

  {{- if .IsURL -}}
    {{- $url := index .Params "url" -}}
    {{- $archive_url := (index .Params "web_archive_url") -}}
    {{- $screenshot_url := (index .Params "web_archive_screenshot_url") -}}

    <a href="{{- $url -}}">{{- .Text -}}</a>

    {{- if $archive_url -}}
      {{- indent 1 "" -}}
      <a href="{{- $archive_url -}}">[archive]</a>
    {{- end -}}

    {{- if $screenshot_url -}}
      {{- indent 1 "" -}}
      <a href="{{- $screenshot_url -}}">[screenshot]</a>
    {{- end -}}
  {{- end -}}

  {{- if .IsImage -}}
    {{- $url := index .Params "url" -}}
    {{- $caption := index .Params "caption" -}}
    <figure><img src="{{- $url -}}"></img><figcaption>
    {{- template "message.entity" $caption -}}
      {{- range $caption -}}
        {{- template "message.entity" . -}}
      {{- end -}}
    </figcaption></figure>
  {{- end -}}

  {{- if or .IsVideo .IsAudio -}}
    {{- $url := index .Params "url" -}}
    {{- $caption := index .Params "caption" -}}
    <figure><video src="{{- $url -}}"></video><figcaption>
    {{- template "message.entity" $caption -}}
      {{- range $caption -}}
        {{- template "message.entity" . -}}
      {{- end -}}
    </figcaption></figure>
  {{- end -}}

  {{- if .IsFile -}}
    {{- $url := index .Params "url" -}}
    {{- if $url -}}
      {{- $filename := index .Params "filename" -}}
      {{- $caption := index .Params "caption" -}}

      <a href="{{- $url -}}">[File]
      {{- if $filename -}}
        {{- $filename | indent 1 -}}
      {{- end -}}
      </a>
      {{- if $caption -}}
        {{- nindent 0 "" -}}
        {{- range $caption -}}
          {{- template "message.entity" . -}}
        {{- end -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{- define "message.replied.entity" -}}
  {{- if .IsText -}}{{- .Text -}}{{- end -}}
  {{- if .IsBold -}}<strong>{{- .Text -}}</strong>{{- end -}}
  {{- if .IsItalic -}}<em>{{- .Text -}}</em>{{- end -}}
  {{- if .IsStrikethrough -}}<del>{{- .Text -}}</del>{{- end -}}
  {{- if .IsUnderline -}}<u>{{- .Text -}}</u>{{- end -}}
  {{- if .IsPre -}}<pre>{{- .Text -}}</pre>{{- end -}}
  {{- if .IsCode -}}<code>{{- .Text -}}</code>{{- end -}}

  {{- /* we used blockquote to format reply message, inner blockquote will break format */ -}}
  {{- /* if .IsBlockquote -}}<blockquote>{{- .Text -}}</blockquote>{{- end */ -}}

  {{- if .IsEmail -}}<a href="mailto:{{- .Text -}}">{{- .Text -}}</a>{{- end -}}
  {{- if .IsPhoneNumber -}}<a href="tel:{{- .Text -}}">{{- .Text -}}</a>{{- end -}}

  {{- if .IsURL -}}
    {{- $url := index .Params "url" -}}
    {{- $archive_url := (index .Params "web_archive_url") -}}
    {{- $screenshot_url := (index .Params "web_archive_screenshot_url") -}}

    <a href="{{- $url -}}">{{- .Text -}}</a>

    {{- if $archive_url -}}
      {{- indent 1 "" -}}
      <a href="{{- $archive_url -}}">[archive]</a>
    {{- end -}}

    {{- if $screenshot_url -}}
      {{- indent 1 "" -}}
      <a href="{{- $screenshot_url -}}">[screenshot]</a>
    {{- end -}}
  {{- end -}}

  {{- if .IsImage -}}
    {{- $url := index .Params "url" -}}
    {{- $caption := index .Params "caption" -}}
    <figure><img src="{{- $url -}}"></img><figcaption>
    {{- template "message.entity" $caption -}}
      {{- range $caption -}}
        {{- template "message.entity" . -}}
      {{- end -}}
    </figcaption></figure>
  {{- end -}}

  {{- if or .IsVideo .IsAudio -}}
    {{- $url := index .Params "url" -}}
    {{- $caption := index .Params "caption" -}}
    <figure><video src="{{- $url -}}"></video><figcaption>
    {{- template "message.entity" $caption -}}
      {{- range $caption -}}
        {{- template "message.entity" . -}}
      {{- end -}}
    </figcaption></figure>
  {{- end -}}

  {{- if .IsFile -}}
    {{- $url := index .Params "url" -}}
    {{- if $url -}}
      {{- $filename := index .Params "filename" -}}
      {{- $caption := index .Params "caption" -}}

      <a href="{{- $url -}}">[File]
      {{- if $filename -}}
        {{- $filename | indent 1 -}}
      {{- end -}}
      </a>
      {{- if $caption -}}
        {{- nindent 0 "" -}}
        {{- range $caption -}}
          {{- template "message.entity" . -}}
        {{- end -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
