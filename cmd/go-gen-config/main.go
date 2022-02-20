package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/evald24/go-gen-config/internal/generator"
)

var (
	templatePath string
	outputPath   string
	configPath   string
)

func main() {
	flag.StringVar(&templatePath, "f", "", "Path to the configuration template file (yaml)")
	flag.StringVar(&outputPath, "o", "", "Path to the generated output file (go)")
	flag.StringVar(&configPath, "c", "", "Path to the generate config file (yaml)")
	flag.Parse()

	if templatePath == "" {
		log.Fatal("template path is empty")
	}

	gen := generator.New(templatePath, outputPath)
	if err := gen.Generate(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Generate template done")
}
