{{- define "freebie-cloudflare.fullname" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "freebie-cloudflare.labels" -}}
app.kubernetes.io/name: freebie-cloudflare
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "freebie-cloudflare.selectorLabels" -}}
app.kubernetes.io/name: freebie-cloudflare
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
