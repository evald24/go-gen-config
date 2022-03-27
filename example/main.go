package main

import (
	"fmt"
	"log"

	"github.com/evald24/go-gen-config/example/config"
)

func main() {
	if err := config.Init("./config.yaml"); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("debug: %+v\n", config.GetDebug())
	fmt.Printf("logLevel %+v\n", config.GetLogLevel())
	fmt.Printf("name %+v\n", config.GetName())
	fmt.Printf("age %+v\n", config.GetAge())

}
