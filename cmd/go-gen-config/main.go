package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/evald24/go-gen-config/internal/generator"
)

var (
	templatePath string
	outputPath   string
	configPath   string
)

func main() {
	flag.StringVar(&templatePath, "t", "", "Path to the configuration template file (yaml)")
	flag.StringVar(&outputPath, "o", "", "Path to the generated output file (go)")
	flag.StringVar(&configPath, "c", "", "Path to the generate config file (yaml)")
	flag.Parse()

	checkNoEmpty(map[string]string{
		"template": templatePath,
		"ouput":    outputPath,
	})

	gen := generator.New(templatePath, outputPath, configPath)
	if err := gen.Generate(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Template generation is complete")
}

func checkNoEmpty(paths map[string]string) {
	for k, v := range paths {
		if strings.TrimSpace(v) == "" {
			log.Fatalf("%v path is empty", k)
		}
	}
}
