package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evald24/go-gen-config/example/config" // The path to the package in your project
)

func main() {
	if err := config.Init("example/config.yaml"); err != nil {
		log.Fatal(err)
	}

	go hotReloadConfig()

	fmt.Printf("config: %+v\n", config.GetConfig())
	fmt.Printf("project name: %s", config.GetConfig().Project.Name)
}

// Example of a hot reload configuration
func hotReloadConfig() {
	signalHotReload := make(chan os.Signal, 1)
	signal.Notify(signalHotReload, syscall.SIGHUP)

	for {
		<-signalHotReload
		if err := config.UpdateConfig(); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("hot reloaded config: %+v\n", time.Now())
	}
}
