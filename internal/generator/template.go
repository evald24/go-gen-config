package generator

import "text/template"

var TemplateConfig = template.Must(template.New("").Parse(`
// Code generated by "go-gen-config"; DO NOT EDIT.

package config

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/evald24/go-gen-config/pkg/helpers"
	"gopkg.in/yaml.v3"
)

{{ define "get" }}
// Get{{.Name}} - {{if .Description }}{{.Description}}{{else}}...{{end}}
func Get{{.Name}}() {{.Type}} { {{if eq .Type "string"}}if cfg.{{.Name}} == "" {
		return "{{.Default}}"
	}{{else if .IsNumber}}if cfg.{{.Name}} == 0 {
	return {{.Default}}
	}{{end}}{{if .IsEnum}} switch cfg.{{.Name}} { {{range $j, $enum :=.Enums}}
		case "{{$enum.Value}}":
			return {{$j}} {{end}}
		 default:
			 return 0
	}{{else}}
	return cfg.{{.Name}}{{end}}
}
{{ end }}

{{ define "struct" }}
type {{.Type}} struct {
	{{ range .Items }}// {{.Name}} - {{.Description}}
	{{.Name}} {{if .IsEnum}}string{{else}}{{.Type}}{{end}}{{.Tags}} // Default: {{.Default}}
{{ end }}}

{{ end }}

type config struct { {{ range . }}{{if .Description }}// {{.Name}} - {{.Description}}{{end}}
	{{.Name}} {{if .IsEnum}}string{{else}}{{.Type}}{{end}}{{.Tags}} // Default: {{.Default}}
{{ end }}}

{{ range $i, $item := . }} {{if .IsEnum}}
// Type - {{.Description}}
type {{.Type}} = uint8
const ({{ range $j, $enum := .Enums}}{{$enum.Name}} {{if eq $j 0}} {{$item.Type}} = iota {{end}}
{{end}})
{{ end }}

{{ template "get" $item }}

{{if $item.IsStruct }}
{{ template "struct" $item }}
{{ end }}
{{ end }}


var fileConfig string
var cfg *config

func Init(configPath string) error {
	fileConfig = configPath

	if err := UpdateConfig(); err != nil {
		return fmt.Errorf("Configuration initialization failed: %v", err)
	}

	hotReload := make(chan os.Signal, 1)
	signal.Notify(hotReload, syscall.SIGHUP)

	go func() {
		for {
			<-hotReload
			UpdateConfig()
		}
	}()

	return nil
}

// UpdateConfig - Updates the config by rereading.
func UpdateConfig() error {
	file, err := os.Open(fileConfig)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return err
	}

	// read environment and replace
	return helpers.ReadEnvAndSet(cfg)
}

`))

type ConfigItem struct {
	Key         string
	Name        string
	Description string
	Type        string
	Tags        string
	Default     string

	IsNumber bool
	IsEnum   bool
	IsStruct bool
	Env      string
	Enums    []EnumKV
	Items    []ConfigItem
}

type EnumKV struct {
	Name  string
	Value string
}
