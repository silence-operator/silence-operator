{{- define "silence-operator.image" -}}
{{- printf "%s:%s" .Values.image.registry (default (printf "v%s" .Chart.AppVersion) .Values.image.tag) }}
{{- end }}

{{- define "chart.name" -}}
{{- if .Chart }}
  {{- if .Chart.Name }}
    {{- .Chart.Name | trunc 63 | trimSuffix "-" }}
  {{- else if .Values.nameOverride }}
    {{ .Values.nameOverride | trunc 63 | trimSuffix "-" }}
  {{- else }}
    silence-operator
  {{- end }}
{{- else }}
  silence-operator
{{- end }}
{{- end }}


{{- define "chart.labels" -}}
{{- if .Chart.AppVersion -}}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
{{- if .Chart.Version }}
helm.sh/chart: {{ .Chart.Version | quote }}
{{- end }}
app.kubernetes.io/name: {{ include "chart.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}


{{- define "chart.selectorLabels" -}}
app.kubernetes.io/name: {{ include "chart.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
