{{- define "freebie-api.fullname" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "freebie-api.labels" -}}
app.kubernetes.io/name: freebie-api
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "freebie-api.selectorLabels" -}}
app.kubernetes.io/name: freebie-api
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
