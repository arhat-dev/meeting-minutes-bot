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

      {{- if entityIsImage . -}}
        {{- $url := index .Params "url" -}}
        {{- $caption := index .Params "caption" -}}
        <figure><img src="{{- $url -}}"></img><figcaption>{{- $caption -}}</figcaption></figure>
      {{- end -}}

      {{- if or (entityIsVideo .) (entityIsAudio .) -}}
        {{- $url := index .Params "url" -}}
        {{- $caption := index .Params "caption" -}}
        <figure><video src="{{- $url -}}"></video><figcaption>{{- $caption -}}</figcaption></figure>
      {{- end -}}

      {{- if entityIsDocument . -}}
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
            <br class="inline"><blockquote>{{- $caption -}}</blockquote>
          {{- end -}}
        {{- end -}}
      {{- end -}}

    {{- end -}} {{- /* entities */ -}}
    </p><hr>
  {{- end -}} {{- /* Message */ -}}
{{- end -}} {{- /* Page */ -}}
