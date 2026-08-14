{{- range . }}
{{- if ne .Name "github.com/silence-operator/silence-operator" }}
{{ .Name }}	{{ .Version }}	{{ .LicenseName }}	{{ .LicenseURL }}	{{ .LicensePath }}
{{ end }}
{{- end }}
