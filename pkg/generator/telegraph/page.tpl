{{- /*
    Page template to render messsages
    Go to https://telegra.ph/api and find supported tags
*/ -}}

{{- if .Messages -}}
  {{- range .Messages -}}
    <p>
    {{- /* Message Author Info */ -}}

    {{- /* Message Body */ -}}
    {{- range .Entities -}}
      {{- if entityIsText . -}}
        {{ .Text }}
      {{- end -}}

      {{- if entityIsBold . -}}
        <strong>{{ .Text }}</strong>
      {{- end -}}

      {{- if entityIsItalic . -}}
        <em>{{ .Text }}</em>
      {{- end -}}

      {{- if entityIsStrikethrough . -}}
        <del>{{ .Text }}</del>
      {{- end -}}

      {{- if entityIsUnderline . -}}
        <u>{{ .Text }}</u>
      {{- end -}}

      {{- if entityIsPre . -}}
        <pre>{{ .Text }}</pre>
      {{- end -}}

      {{- if entityIsCode . -}}
        <code>{{ .Text }}</code>
      {{- end -}}

      {{- if entityIsNewLine . -}}
        {{ .Text }}<br>
      {{- end -}}

      {{- if entityIsParagraph . -}}
        <p>{{ .Text }}</p>
      {{- end -}}

      {{- if entityIsThematicBreak . -}}
        {{ .Text }}<hr>
      {{- end -}}

      {{- if entityIsBlockquote . -}}
        <blockquote>{{ .Text }}</blockquote>
      {{- end -}}

      {{- if entityIsEmail . -}}
        <a href="mailto:{{ .Text }}">{{ .Text }}</a>
      {{- end -}}

      {{- if entityIsPhoneNumber . -}}
        <a href="tel:{{ .Text }}">{{ .Text }}</a>
      {{- end -}}

      {{- if entityIsURL . -}}
        <a href="{{ index .Params "url" }}">{{ .Text }}</a>
      {{- end -}}

      {{- if entityIsImage . -}}
        <figure><img src="{{ index .Params "url" }}"></img><figcaption>{{ index .Params "caption" }}</figcaption></figure>
      {{- end -}}

      {{- if or (entityIsVideo .) (entityIsAudio .) -}}
        <figure><video src="{{ index .Params "url" }}"></video><figcaption>{{ index .Params "caption" }}</figcaption></figure>
      {{- end -}}

      {{- if entityIsDocument . -}}
        <a href="{{ index .Params "url" }}">[File] {{ index .Params "filename" }} - {{ index .Params "caption" }}</a>
      {{- end -}}
    {{- end -}}
    </p><hr>
  {{- end -}}
{{- end -}}

{{- /*
  cleanup end of the page
*/ -}}
