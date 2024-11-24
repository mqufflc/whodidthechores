{{/*
Expand the name of the chart.
*/}}
{{- define "whodidthechores.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "whodidthechores.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Namespaces deployed in
*/}}
{{- define "whodidthechores.namespace" -}}
{{- default .Release.Namespace .Values.namespace -}}
{{- end -}}

{{/*
Chart Name
*/}}
{{- define "whodidthechores.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Service Account name
*/}}
{{- define "whodidthechores.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "whodidthechores.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Standard Kubernetes labels
*/}}
{{- define "whodidthechores.labels" -}}
app.kubernetes.io/name: {{ include "whodidthechores.name" . }}
helm.sh/chart: {{ include "whodidthechores.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/part-of: whodidthechores
{{- end -}}

{{/*
Labels to use in selectors
*/}}
{{- define "whodidthechores.matchLabels" -}}
app.kubernetes.io/name: {{ include "whodidthechores.name" . }}
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
