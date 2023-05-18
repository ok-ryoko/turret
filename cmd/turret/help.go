package main

import (
	"github.com/urfave/cli/v2"
)

func customizeHelpTemplates() {
	cli.AppHelpTemplate = appHelpTemplate
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.SubcommandHelpTemplate = commandHelpTemplate
}

var appHelpTemplate string = `{{.Name}} v{{.Version}}
{{.Usage}}

Usage: {{.HelpName}} [command]

{{if .Commands -}}
Commands:

{{range .Commands}}  {{join .Names ", " }}{{ "\t\t" }}{{.Usage}}
{{end}}
{{end}}`

var commandHelpTemplate string = `{{.Usage}}

Usage: {{if .UsageText -}}
    {{- .UsageText -}}
  {{- else -}}
    {{- .HelpName -}}
    {{- if .VisibleFlags}} [options]{{end -}}
    {{- if .ArgsUsage}} {{.ArgsUsage}}{{end}}
  {{- end}}

{{if .VisibleFlags -}}
Options:

{{range .VisibleFlags}}  {{ . }}
{{end}}
{{end}}`
