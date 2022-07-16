{{- define "artifacts.bin-name" -}}

mbot-{{- matrix.kernel -}}-{{- matrix.arch -}}

  {{- if eq matrix.kernel "windows" -}}
    .exe
  {{- end -}}

{{- end -}}
