{{/*
Expand the name of the chart.
*/}}
{{- define "whodidthechores.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "whodidthechores.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "whodidthechores.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "whodidthechores.labels" -}}
helm.sh/chart: {{ include "whodidthechores.chart" . }}
{{ include "whodidthechores.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "whodidthechores.selectorLabels" -}}
app.kubernetes.io/name: {{ include "whodidthechores.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "whodidthechores.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "whodidthechores.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Standard Kubernetes labels for pg
*/}}
{{- define "whodidthechores.postgresLabels" -}}
app.kubernetes.io/name: {{ include "whodidthechores.name" . }}-pg
helm.sh/chart: {{ include "whodidthechores.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/part-of: whodidthechores
{{- end -}}

{{/*
Labels to use in selectors for pg
*/}}
{{- define "whodidthechores.postgresMatchLabels" -}}
app.kubernetes.io/name: {{ include "whodidthechores.name" . }}-pg
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Helper to render a template inside a value
*/}}
{{- define "whodidthechores.render" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}
