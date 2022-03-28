package main

import (
	"fmt"
	"log"

	"github.com/evald24/go-gen-config/example/config"
)

func main() {
	if err := config.Init("example/config.yaml"); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("config: %+v\n", config.GetConfig())
	project := config.GetConfig().Project
	fmt.Printf("project name: %s", project.Name)
}
