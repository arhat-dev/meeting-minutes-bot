{{- /*
    Page template to render messsages
    Go to https://telegra.ph/api and find supported tags
*/ -}}

{{- if .Messages -}}
  {{- range .Messages -}}
    {{- if not .IsPrivateMessage -}}
      <p>
      {{- /* Message Link */ -}}
      
      {{- if .MessageURL -}}
        <a href="{{- .MessageURL -}}">[Link]</a>
      {{- end -}}

      {{- /* Message Author Info */ -}}

      {{- if .AuthorURL -}}
        {{- if .MessageURL -}}
          {{- indent 1 "" -}}
        {{- end -}}
        <a href="{{- .AuthorURL -}}">
      {{- end -}}
        {{- .Author -}}
      {{- if .AuthorURL -}}
        </a>
      {{- end -}}

      {{- if .ChatName -}}
        @
        {{- indent 1 "" -}}
        {{- if .ChatURL -}}
          <a href="{{- .ChatURL -}}">
        {{- end -}}

          {{- .ChatName -}}

        {{- if .ChatURL -}}
          </a>
        {{- end -}}
      {{- end -}}

      {{- if and .IsForwarded (gt (len .OriginalAuthor) 0) -}}
        {{- indent 1 "" -}}
        from
        {{- if .OriginalAuthorURL -}}
          <a href="{{- .OriginalAuthorURL -}}">
        {{- end -}}
          {{- .OriginalAuthor -}}
        {{- if .OriginalAuthorURL -}}
          </a>
        {{- end -}}

        {{- if .OriginalChatName -}}
          @
          {{- indent 1 "" -}}
          {{- if .OriginalChatURL -}}
            <a href="{{- .OriginalChatURL -}}">
          {{- end -}}
            {{- .OriginalChatName -}}
          {{- if .OriginalChatURL -}}
            </a>
          {{- end -}}
        {{- end -}}
      {{- end -}}
      </p>
    {{- end -}}

    {{- /* Message Body */ -}}

    <p>
    {{- range .Entities -}}
      {{- if entityIsText . -}}
        {{- .Text -}}
      {{- end -}}

      {{- if entityIsBold . -}}
        <strong>{{- .Text -}}</strong>
      {{- end -}}

      {{- if entityIsItalic . -}}
        <em>{{- .Text -}}</em>
      {{- end -}}

      {{- if entityIsStrikethrough . -}}
        <del>{{- .Text -}}</del>
      {{- end -}}

      {{- if entityIsUnderline . -}}
        <u>{{- .Text -}}</u>
      {{- end -}}

      {{- if entityIsPre . -}}
        <pre>{{- .Text -}}</pre>
      {{- end -}}

      {{- if entityIsCode . -}}
        <code>{{- .Text -}}</code>
      {{- end -}}

      {{- if entityIsThematicBreak . -}}
        {{- .Text -}}<hr>
      {{- end -}}

      {{- if entityIsBlockquote . -}}
        <blockquote>{{- .Text -}}</blockquote>
      {{- end -}}

      {{- if entityIsEmail . -}}
        <a href="mailto:{{- .Text -}}">{{- .Text -}}</a>
      {{- end -}}

      {{- if entityIsPhoneNumber . -}}
        <a href="tel:{{- .Text -}}">{{- .Text -}}</a>
      {{- end -}}

      {{- if entityIsURL . -}}
        <a href="{{- index .Params "url" -}}">{{- .Text -}}</a>
      {{- end -}}

      {{- if entityIsImage . -}}
        <figure><img src="{{- index .Params "url" -}}"></img><figcaption>{{- index .Params "caption" -}}</figcaption></figure>
      {{- end -}}

      {{- if or (entityIsVideo .) (entityIsAudio .) -}}
        <figure><video src="{{- index .Params "url" -}}"></video><figcaption>{{- index .Params "caption" -}}</figcaption></figure>
      {{- end -}}

      {{- if and (entityIsDocument .) (gt (len (index .Params "url")) 0) -}}
        <a href="{{- index .Params "url" -}}">[File]
        {{- $filename := index .Params "filename" -}}
        {{- if $filename -}}
          {{- $filename | indent 1 -}}
        {{- end -}}

        {{- if (index .Params "caption") -}}
          {{- if $filename -}}
            {{- indent 1 "-" -}}
          {{- else -}}
            {{- indent 1 " " -}}
          {{- end -}}
          {{- index .Params "caption" -}}
        {{- end -}}
        </a>
      {{- end -}}

    {{- end -}} {{- /* entities */ -}}
    </p><hr>
  {{- end -}} {{- /* Message */ -}}
{{- end -}} {{- /* Page */ -}}
