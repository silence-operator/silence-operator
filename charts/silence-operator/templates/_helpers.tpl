{{- define "silence-operator.image" -}}
{{- printf "%s:%s" .Values.image.registry (default (printf "v%s" .Chart.AppVersion) .Values.image.tag) }}
{{- end }}
