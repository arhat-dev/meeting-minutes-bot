{{-
/*
    Page footer template to append after each generation
*/ -}}

{{- define "page.footer" -}}
{{- $now := now -}}
<aside>generated at {{ $now.UTC.Format "2006-01-02T15:04:05Z07:00"  }}</aside><hr>
{{- end -}}
