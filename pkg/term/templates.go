package term

var podDetailsTemplate = `Name:          {{ .Name }}
Namespace:     {{ .Namespace }}
Node:          {{ .Spec.NodeName }}
{{- if .OwnerReferences }}
{{- $owner := index .OwnerReferences 0 }}
Controlled By: {{ $owner.Kind }}/{{ $owner.Name }}
{{- end }}
Created:       {{ .CreationTimestamp }}

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
