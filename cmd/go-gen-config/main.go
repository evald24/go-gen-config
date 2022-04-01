package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/evald24/go-gen-config/internal/generator"
)

var (
	templatePath string
	outputPath   string
	configPath   string
)

var RootCmd = &cobra.Command{
	Use:     "go-gen-config",
	Short:   "Template-based configuration generator",
	Version: "v0.4.0",
	Run:     run,
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&templatePath, "template", "t", "", "path to the configuration template file (yaml)")
	RootCmd.PersistentFlags().StringVarP(&outputPath, "output", "o", "", "path to the generated output file (go)")
	RootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "path to the generate config file (yaml)")

	if err := RootCmd.MarkPersistentFlagRequired("template"); err != nil {
		fatal(err)
	}

	if err := RootCmd.MarkPersistentFlagRequired("output"); err != nil {
		fatal(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	if strings.TrimSpace(templatePath) == "" {
		fatal("template path is empty")
	}

	if strings.TrimSpace(outputPath) == "" {
		fatal("ouput path is empty")
	}

	gen := generator.New(templatePath, outputPath, configPath)
	if err := gen.Generate(); err != nil {
		fatal(err)
	}

	fmt.Println("Template generation is complete")
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func fatal(v ...interface{}) {
	fmt.Println(v...)
	os.Exit(1)
}
