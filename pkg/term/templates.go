package term

const podDetailsTemplate = `Name:          {{ .Name }}
Namespace:     {{ .Namespace }}
Node:          {{ .Spec.NodeName }}
{{- if .OwnerReferences }}
{{- range $idx, $owner := .OwnerReferences }}
Controlled By: {{ $owner.Kind }}/{{ $owner.Name }}
{{- end }}
{{- end }}
Created:       {{ .CreationTimestamp }}
{{- if .Status.ContainerStatuses }}
{{- $cont := index .Status.ContainerStatuses 0 }}
Restarts:      {{ $cont.RestartCount }}
{{- end }}

Status:        {{ .Status.Phase }}
IP:            {{ .Status.PodIP }}

Containers:
{{- range $idx, $container := .Spec.Containers }}

   {{ $container.Name }}:
       Image: {{ $container.Image }}
       Environment:
         {{ range $k, $v := $container.Env -}}
           {{ $v.Name }}: {{ $v.Value }}
         {{ end -}}
{{ end -}}
`
