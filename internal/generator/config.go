package generator

import "text/template"

// debug: true
// logLevel: INFO
// logFile: ${HOME}/logs.log

// project:
//   name: my-service-api
//   title: My Service API
//   description: Service description
//   environment: STG

// grpc:
//   host: 127.0.0.1
//   port: 50051
//   maxConnectionIdle: 5 # Minutes
//   maxConnectionAge: 5 # Minutes
//   maxConnectionAgeGrace: 5 # Minutes
//   time: 15 # Minutes
//   timeout: 15 # Seconds

var TemplateConfig = template.Must(template.New("").Parse(`{{ range . }}
{{.Key}}: {{.Default}} {{if .Description}}# {{.Description}}{{end}}
{{ end }}
`))