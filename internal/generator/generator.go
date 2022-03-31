package generator

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

func getParams(cfgMap map[string]interface{}, structName string) (map[string]ConfigItem, error) {

	params := make(map[string]ConfigItem)
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

		// Optimizing structures for efficient memory allocation
		var key string
		switch valueType {
		case "bool":
			key = "0"
		case "uint8", "byte", "int8":
			key = "1"
		case "uint16", "int16":
			key = "2"
		case "uint32", "int32", "rune", "int", "uint", "float32":
			key = "3"
		case "uint64", "int64", "float64":
			key = "4"
		case "string":
			key = "7"
		case "enum":
			key = "8"
		default:
			key = "9"
		}
		key += valueType + k

		tags := fmt.Sprintf("yaml:\"%s\"", k)
		if len(env) > 0 {
			tags += fmt.Sprintf(" env:\"%s\"", env)
		}
		if defaultValue != "" && valueType != "struct" {
			tags += fmt.Sprintf(" default:\"%s\"", defaultValue)
		}

		tags = fmt.Sprintf("`%s`", tags)

		if helpers.Contains(baseTypes, valueType) {
			params[key] = ConfigItem{
				Name:        name,
				Description: description,
				Type:        valueType,
				Tags:        tags,
				Env:         env,
				Default:     defaultValue,
			}
			continue
		}

		if valueType == "enum" {
			enums := helpers.GetEnum(value)
			enumKV := make([]EnumKV, 0, len(enums))

			for _, v := range enums {
				enumKV = append(enumKV, EnumKV{
					Name:  strcase.ToCamel(structName + "_" + name + "_" + strings.ToLower(v)),
					Value: v,
				})
			}

			params[key] = ConfigItem{
				Name:        name,
				Description: description,
				Type:        strcase.ToCamel(valueType + "_" + structName + "_" + name),
				Tags:        tags,
				IsEnum:      true,
				Enums:       enumKV,
				Env:         env,
				Default:     defaultValue,
			}
			continue
		}

		if valueType == "struct" {
			items, err := getParams(value["value"].(map[string]interface{}), name)
			if err != nil {
				return nil, err
			}

			params[key] = ConfigItem{
				Name:        name,
				Description: description,
				Type:        strcase.ToCamel(valueType + "_" + structName + "_" + name),
				Tags:        tags,
				IsStruct:    true,
				Env:         env,
				Items:       items,
			}
			continue
		}

		return nil, fmt.Errorf("failed to generate code, unsupported type \"%s\" for the \"%s\" field", valueType, k)
	}

	return params, nil
}

func (g *generator) buildTemplate(tpl *template.Template, params map[string]ConfigItem) (*bytes.Buffer, error) {
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

	params, err := getParams(g.cfgMap, "")
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

	// read or create a configuration file
	configFile, _ := os.OpenFile(g.configPath, os.O_RDWR, 0770)
	if configFile != nil {
		var oldConfig map[string]interface{}
		if err := yaml.NewDecoder(configFile).Decode(&oldConfig); err == nil {
			config = mergeDiffConfig(config, oldConfig)
		}
	} else {
		configFile, err = helpers.CreateFile(g.configPath)
		if err != nil {
			return fmt.Errorf("create config file: %v", err)
		}
	}
	defer configFile.Close()

	if err := os.Truncate(g.configPath, 0); err != nil {
		return fmt.Errorf("create config file: %v", err)
	}

	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)

	if err := yamlEncoder.Encode(&config); err != nil {
		return err
	}

	if _, err := configFile.WriteAt(b.Bytes(), 0); err != nil {
		return err
	}

	// Formatting the code

	dir, err := helpers.GoRoot()
	if err != nil {
		return err
	}

	cmd := exec.Command(filepath.Join(dir, "bin/gofmt"), "-s", "-w", g.outputPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start gofmt: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to execute gofmt: %v", err)
	}

	return nil
}

// mergeDiffConfig - replace the new value with the old one
func mergeDiffConfig(new, old map[string]interface{}) map[string]interface{} {
	for k, v := range old {
		if v, ok := v.(map[string]interface{}); ok {
			new[k] = mergeDiffConfig(new[k].(map[string]interface{}), v)
			continue
		}
		new[k] = v
	}

	return new
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
