{{- define "name" -}}
{{- .Chart.Name -}}
{{- end -}}

{{- define "fullname" -}}
{{- .Chart.Name -}}
{{- end -}}

{{- define "domain" -}}
{{- printf "%s.%s.%s" .Values.ecc.sub_domain .Release.Namespace .Values.domain -}}
{{- end -}}

{{- define "domain-api" -}}
{{- printf "%s.%s.%s" .Values.api.sub_domain .Release.Namespace .Values.domain -}}
{{- end -}}

{{- define "serviceAccountName" -}}
	{{ default "default" .Values.serviceAccount.name }}
{{- end -}}