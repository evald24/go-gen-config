package generator

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/evald24/go-gen-config/internal/helpers"
	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v3"
)

type generator struct {
	templatePath string
	outputPath   string
	configPath   string
	cfgMap       map[string]interface{}
}

func New(templatePath, outputPath, configPath string) *generator {
	return &generator{
		templatePath: templatePath,
		outputPath:   outputPath,
		configPath:   configPath,
	}
}

func (g *generator) readTemplate() error {
	file, err := os.Open(g.templatePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&g.cfgMap); err != nil {
		return err
	}

	return nil
}

func getParams(cfgMap map[string]interface{}) ([]ConfigItem, error) {

	params := make([]ConfigItem, 0, len(cfgMap))
	for k, v := range cfgMap {
		value, ok := v.(map[string]interface{})
		if !ok {
			log.Fatal("value must be a map[string]interface{}")
		}

		valueType := helpers.GetString(value, "type")
		description := helpers.GetString(value, "description")
		name := strcase.ToCamel(k)
		defaultValue := helpers.GetString(value, "value")
		env := helpers.GetString(value, "env")

		envTag := ""
		if len(env) > 0 {
			envTag = fmt.Sprintf(" env:\"%s\"", env)
		}

		tags := fmt.Sprintf("`yaml:\"%s\"%s`", k, envTag)

		if helpers.Contains(baseTypes, valueType) {
			params = append(params, ConfigItem{
				Key:         k,
				Name:        name,
				Description: description,
				Type:        valueType,
				Tags:        tags,
				Env:         env,
				Default:     defaultValue,
			})
		}

		if valueType == "enum" {
			constType := strcase.ToCamel(fmt.Sprintf("%v_%v", valueType, name))
			enums := helpers.GetEnum(value)
			enumKV := make([]EnumKV, 0, len(enums))

			for _, v := range enums {
				enumKV = append(enumKV, EnumKV{
					Name:  strcase.ToCamel(name + "_" + strings.ToLower(v)),
					Value: v,
				})
			}

			params = append(params, ConfigItem{
				Key:         k,
				Name:        name,
				Description: description,
				Type:        constType,
				Tags:        tags,
				IsEnum:      true,
				Enums:       enumKV,
				Env:         env,
				Default:     defaultValue,
			})
		}

		if valueType == "struct" {
			constType := strcase.ToCamel(fmt.Sprintf("%v_%v", valueType, name))

			items, err := getParams(value["value"].(map[string]interface{}))
			if err != nil {
				return nil, err
			}

			params = append(params, ConfigItem{
				Key:         k,
				Name:        name,
				Description: description,
				Type:        constType,
				Tags:        tags,
				IsStruct:    true,
				Env:         env,
				Default:     defaultValue,
				Items:       items,
			})
		}
	}

	return params, nil
}

func (g *generator) buildTemplate(tpl *template.Template, params []ConfigItem) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, params); err != nil {
		return nil, fmt.Errorf("execute template: %v", err)
	}

	return &buf, nil
}

var baseTypes = []string{
	"int", "int8", "int16", "int32", "int64", "rune",
	"uint", "uint8", "uint16", "uint32", "uint64", "byte",
	"string", "bool", "float32", "float64",
}

func (g *generator) Generate() error {
	if err := g.readTemplate(); err != nil {
		return err
	}

	params, err := getParams(g.cfgMap)
	if err != nil {
		return err
	}

	// Generate code
	bufCode, err := g.buildTemplate(TemplateConfig, params)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	astOutFile, err := parser.ParseFile(fset, "", bufCode.Bytes(), parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse template: %v", err)
	}

	outFile, err := helpers.CreateFile(g.outputPath)
	if err != nil {
		return fmt.Errorf("create file: %v", err)
	}
	defer outFile.Close()

	err = printer.Fprint(outFile, fset, astOutFile)
	if err != nil {
		log.Fatalf("print file: %v", err)
	}

	// Generate config file

	if g.configPath == "" {
		return nil
	}

	config := getConfig(g.cfgMap)

	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	configFile, err := helpers.CreateFile(g.configPath)
	if err != nil {
		return fmt.Errorf("create config file: %v", err)
	}
	defer configFile.Close()

	if _, err := configFile.Write(bytes); err != nil {
		return err
	}

	return nil
}

func getConfig(cfg map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range cfg {
		value, _ := v.(map[string]interface{})
		if value["type"] == "struct" {
			result[k] = getConfig(value["value"].(map[string]interface{}))
		} else if value["value"] != nil {
			result[k] = value["value"]
		}
	}

	return result
}
