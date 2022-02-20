package main

import (
	"fmt"

	config "github.com/evald24/go-gen-config"
)

func main() {
	config.Init("./config.yaml")
	fmt.Printf("debug: %+v\n", config.GetDebug())
	fmt.Printf("logLevel %+v\n", config.GetLogLevel())
	fmt.Printf("name %+v\n", config.GetName())
	fmt.Printf("age %+v\n", config.GetAge())

}
