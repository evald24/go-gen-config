# Go generate configuration

![License](https://img.shields.io/github/license/evald24/go-gen-config?color=blue)
![Release](https://img.shields.io/github/v/release/evald24/go-gen-config)
![Code size](https://img.shields.io/github/languages/code-size/evald24/go-gen-config?color=blue)
![Go version](https://img.shields.io/github/go-mod/go-version/evald24/go-gen-config?color=blue&logo=go)

A convenient configuration generator based on an `yaml` template with `env` support and the possibility of a hot reload

## Install

```zsh
GOPATH=./bin go install github.com/evald24/go-gen-config/cmd/go-gen-config
```

## Usage:

At the beginning, create a file with a configuration template (e.g. [`config.template.yaml`](/example/config.template.yaml))

Then execute the generation command:

```zsh
./bin/go-gen-config -t ./example/config.template.yaml -o ./example/config/config.gen.go -c ./example/config.yaml
```

| Args | Description                                                                                                                      |
| :--: | -------------------------------------------------------------------------------------------------------------------------------- |
| `-t` | `required` The path to the yaml template file                                                                                    |
| `-o` | `required` The path where the code will be generated                                                                             |
| `-c` | `optional` The path where the configuration file will be generated. If the file already exists, it only adds new default values. |

## Schema

### Template

```yaml
appName: # The name will be converted to CamelCase
  description: Application Name # Its description
  type: string # The type in which the parameter is specified
  env: APP_NAME # A unique value that is set from the environment
  value: Hello World # The value in the specified data type

logLevel:
  type: enum
  env: LOG_LEVEL
  value: DEBUG
  enum:
    - DEBUG
    - INFO
    - WARNING
    - ERROR

project:
  type: struct
  description: Project information
  value:
    name:
      type: string
      value: example-service-api
      description: Name of the project
    title:
      type: string
      value: Example of a service API
      description: Title of the project
    description:
      type: string
      value: Description for an example of a service API
      description: Description of the project
    environment:
      type: enum
      value: DEV
      env: PROJECT_ENV
      enum: [DEV, STG, PROD]
      description: Project environment
```

### Outputs

#### Generated configuration file

If the `-c` argument was specified

```yaml
appName: Hello World
logLevel: DEBUG
project:
  description: Description for an example of a service API
  environment: DEV
  name: example-service-api
  title: Example of a service API
```

#### Code generated

```go
// Code generated by "go-gen-config"; DO NOT EDIT.

package config

...

// Conifg - Basic structure with configuration
type Config struct {
	// Debug - Debug mode
	Debug bool `yaml:"debug" env:"DEBUG" default:"true"`
	// LogLevel - Description of the variable
	LogLevel EnumLogLevel `yaml:"logLevel" env:"LOG_LEVEL" default:"DEBUG"`
	// Project - Project information
	Project StructProject `yaml:"project"`
}

// EnumLogLevel - Description of the variable
type EnumLogLevel = string

const (
	// LogLevelDebug - Description of the variable
	LogLevelDebug EnumLogLevel = "DEBUG"
	// LogLevelInfo - Description of the variable
	LogLevelInfo = "INFO"
	// LogLevelWarning - Description of the variable
	LogLevelWarning = "WARNING"
	// LogLevelError - Description of the variable
	LogLevelError = "ERROR"
)

// StructProject - Project information
type StructProject struct {
	// Description - Description of the project
	Description string `yaml:"description" default:"Description for an example of a service API"`
	// Name - Name of the project
	Name string `yaml:"name" default:"example-service-api"`
	// Title - Title of the project
	Title string `yaml:"title" default:"Example of a service API"`
	// Environment - Project environment
	Environment EnumProjectEnvironment `yaml:"environment" env:"PROJECT_ENV" default:"DEV"`
}

// EnumProjectEnvironment - Project environment
type EnumProjectEnvironment = string

const (
	// ProjectEnvironmentDev - Project environment
	ProjectEnvironmentDev EnumProjectEnvironment = "DEV"
	// ProjectEnvironmentStg - Project environment
	ProjectEnvironmentStg = "STG"
	// ProjectEnvironmentProd - Project environment
	ProjectEnvironmentProd = "PROD"
)

// GetConfig - get the configuration
func GetConfig() *Config {
	return cfg
}

var fileConfig string
var cfg *Config

// Init - initializing the configuration
func Init(configPath string) (*Config, error) {
	...
}

// UpdateConfig - Updates the configuration by rereading
func UpdateConfig() error {
	...
}

```

#### Usage Example

```go
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
	cfg, err := config.Init("example/config.yaml");
	if err != nil {
		log.Fatal(err)
	}

	go hotReloadConfig()

	fmt.Printf("config: %+v\n", cfg)
	fmt.Printf("project name: %s", cfg.Project.Name)
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
```

If you need to update the configuration on the fly, you can send a `SIGHUP' signal to the process.

```zsh
kill -SIGHUP [PID]
```

The configuration file and the environment will be read again

### Types

You can see the full example of the template here: [`example/config.template.yaml`](./example/config.template.yaml)

| Attribute     | Description                                                        |
| ------------- | ------------------------------------------------------------------ |
| `type`        | `required` The type in which the parameter is specified            |
| `value`       | `optional` Default value                                           |
| `env`         | `optional` The value in the specified data type                    |
| `description` | `optional` Description, will be converted to a comment in the code |
| `enum`        | `optional` Array of enumerations                                   |

### Support types

| Attribute                                             | Description                                                                                                             |
| ----------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `bool`                                                | Boolean values are those which can be assigned `true` or `false`                                                        |
| `string`                                              | Text, the type will be specified in the code similarly                                                                  |
| `enum`                                                | Define an enumeration as a list of enumeration elements in the `enum` field                                             |
| `int`, `int8`, `int16`, `int32`, `int64`, `rune`      | Integer, the type will be specified in the code similarly                                                               |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `byte` | A unsigned integer, the type will be specified in the code similarly                                                    |
| `float32`, `float64`                                  | Single precision floating point format, the type will be specified in the code similarly                                |
| `struct`                                              | The structure in which the fields are specified in `value` can be used for grouping. Can recursively contain structures |
