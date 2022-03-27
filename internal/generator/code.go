package generator

import "text/template"

var TemplateCode = template.Must(template.New("").Parse(`
// Code generated by "go-gen-config"; DO NOT EDIT.

package config

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

type config struct { {{ range . }}
	{{if .Description }}// {{.Name}} - {{.Description}}{{end}}
	{{.Name}} {{if .IsEnum}}string{{else}}{{.Type}}{{end}}{{.Tags}} // Default: {{.Default}}
{{ end }}}

{{ range $i, $item := . }} {{if .IsEnum}}
// Enum {{.Name}}
type {{.Type}} = uint8
const ({{ range $j, $enum := .Enums}}
	{{$enum.Name}} {{if eq $j 0}} {{$item.Type}} = iota {{end}}
{{end}}){{ end }}

// Get{{.Name}} - {{if .Description }}{{.Description}}{{else}}...{{end}}
func Get{{.Name}}() {{.Type}} {
	{{if eq .Type "string"}}if cfg.{{.Name}} == "" {
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
			fmt.Printf("Hot reloading configuration ...\n")
			if err := UpdateConfig(); err != nil {
				fmt.Printf("Error when hot-reloading the configuration: %v\n", err)
				return
			}
			fmt.Printf("Config hot reloading was successful\n")
		}
	}()

	fmt.Printf("Config initialization was successful\n")
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
	readEnvAndSet(reflect.ValueOf(cfg))
	return nil
}

// readEnvAndSet - Sets config from environment values.
func readEnvAndSet(v reflect.Value) {
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type.Kind() == reflect.Struct {
			readEnvAndSet(v.Field(i))
		} else if tag := field.Tag.Get("env"); tag != "" {
			if value := os.Getenv(tag); value != "" {
				if err := setValue(v.Field(i), value); err != nil {
					fmt.Printf("Failed to set environment value for \"%s\"", field.Name)
				}
			}
		}
	}
}

func setValue(field reflect.Value, value string) error {
	valueType := field.Type()
	switch valueType.Kind() {
	// set string value
	case reflect.String:
		field.SetString(value)

	// set boolean value
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)

	// set integer (or time) value
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Kind() == reflect.Int64 && valueType.PkgPath() == "time" && valueType.Name() == "Duration" {
			// try to parse time
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
		} else {
			// parse regular integer
			number, err := strconv.ParseInt(value, 0, valueType.Bits())
			if err != nil {
				return err
			}
			field.SetInt(number)
		}

	// set unsigned integer value
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		number, err := strconv.ParseUint(value, 0, valueType.Bits())
		if err != nil {
			return err
		}
		field.SetUint(number)

	// set floating point value
	case reflect.Float32, reflect.Float64:
		number, err := strconv.ParseFloat(value, valueType.Bits())
		if err != nil {
			return err
		}
		field.SetFloat(number)

	// unsupported types
	case reflect.Map, reflect.Ptr,
		reflect.Complex64, reflect.Interface,
		reflect.Invalid, reflect.Slice, reflect.Func,
		reflect.Array, reflect.Chan, reflect.Complex128,
		reflect.Struct, reflect.Uintptr, reflect.UnsafePointer:
	default:
		return fmt.Errorf("unsupported type: %v", valueType.Kind())
	}

	return nil
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
}

type EnumKV struct {
	Name  string
	Value string
}