{{- define "freebie-cronjobs.fullname" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "freebie-cronjobs.labels" -}}
app.kubernetes.io/name: freebie-cronjobs
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}
