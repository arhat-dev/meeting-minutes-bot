{{- define "artifacts.bin-name" -}}

app-{{- matrix.kernel -}}-{{- matrix.arch -}}

  {{- if eq matrix.kernel "windows" -}}
    .exe
  {{- end -}}

{{- end -}}
