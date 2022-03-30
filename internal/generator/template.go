package generator

import "text/template"

var TemplateConfig = template.Must(template.New("").Parse(`
// Code generated by "go-gen-config"; DO NOT EDIT.

package config

import (
	"fmt"
	"os"

	"github.com/evald24/go-gen-config/pkg/helpers"
	"gopkg.in/yaml.v3"
)

{{define "struct_item"}}{{.Name}} {{if .IsEnum}}{{.Type}}{{else}}{{.Type}}{{end}}{{.Tags}}{{end}}

// Conifg - Basic structure with configuration
type Config struct {
{{range .}}  // {{.Name}} - {{.Description}}
	{{template "struct_item" .}}
{{end}}}



{{define "enum"}}{{$item := .}}
	// {{.Type}} - {{.Description}}
	type {{.Type}} = string
	const (
		{{range $j, $enum := .Enums}}// {{.Name}} - {{$item.Description}}
		{{$enum.Name}} {{if eq $j 0}} {{$item.Type}} = "{{$enum.Value}}"{{else}} = "{{$enum.Value}}"{{end}}
	{{end}})
{{end}}

{{define "struct"}}
	// {{.Type}} - {{.Description}}
	type {{.Type}} struct {
	{{range .Items}}// {{.Name}} - {{.Description}}
		{{template "struct_item" .}}
	{{end}}}

	{{template "gen" .Items}}
{{end}}

{{define "gen"}}
	{{range .}}
		{{if .IsEnum}}
			{{template "enum" .}}
		{{end}}
		{{if .IsStruct}}
			{{template "struct" .}}
		{{end}}
	{{end}}
{{end}}

{{template "gen" .}}

// GetConfig - get the configuration
func GetConfig() Config {
	return *cfg
}

var fileConfig string
var cfg *Config

// Init - initializing the configuration
func Init(configPath string) error {
	fileConfig = configPath

	if cfg != nil {
		return fmt.Errorf("The configuration has already been initialized")
	}

	if err := UpdateConfig(); err != nil {
		return fmt.Errorf("Configuration initialization failed: %v", err)
	}

	return nil
}

// UpdateConfig - Updates the configuration by rereading
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
