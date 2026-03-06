{{- define "freebie-scheduler.fullname" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "freebie-scheduler.labels" -}}
app.kubernetes.io/name: freebie-scheduler
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}
