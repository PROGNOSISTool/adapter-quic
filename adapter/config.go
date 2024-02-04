package adapter

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
    AdapterAddress string `yaml:"adapterAddress"`
    SulAddress string `yaml:"sulAddress"`
    SulName string `yaml:"sulName"`
    HTTP3 bool `yaml:"HTTP3"`
    HttpPath string `yaml:"httpPath"`
    Tracing  bool          `yaml:"tracing"`
    WaitTime time.Duration `yaml:"WaitTime"`
}

func newConfig() Config {
    waitTime, _ := time.ParseDuration("300ms")
    c := Config{
        AdapterAddress: "0.0.0.0:3333",
        SulAddress:     "implementation:4433",
        SulName:        "quic.tiferrei.com",
        HTTP3:          false,
        HttpPath:       "/index.html",
        Tracing:        false,
        WaitTime:       waitTime,
    }

    return c
}

func GetConfig(path string) *Config {
    config := newConfig()

    yamlFile, fileErr := ioutil.ReadFile(path)
    if fileErr == nil {
        type aliasAdapterConfig struct {
            AdapterAddress string `yaml:"adapterAddress"`
            SulAddress string     `yaml:"sulAddress"`
            SulName string        `yaml:"sulName"`
            HTTP3 bool            `yaml:"http3"`
            HttpPath string       `yaml:"httpPath"`
            Tracing  bool         `yaml:"tracing"`
            WaitTime string       `yaml:"waitTime"`
        }

        type aliasConfig struct {
            Adapter aliasAdapterConfig `yaml:"adapter"`
        }

        alias := aliasConfig{}
        yamlErr := yaml.Unmarshal(yamlFile, &alias)
        if yamlErr != nil {
            fmt.Printf("Falied to unmarshal YAML: %v\n", yamlErr)
        } else {
            config.AdapterAddress = alias.Adapter.AdapterAddress
            config.SulAddress = alias.Adapter.SulAddress
            config.SulName = alias.Adapter.SulName
            config.HTTP3 = alias.Adapter.HTTP3
            config.HttpPath = alias.Adapter.HttpPath
            config.Tracing = alias.Adapter.Tracing

            waitTime, err := time.ParseDuration(alias.Adapter.WaitTime)
            if err == nil {
                config.WaitTime = waitTime
            }
        }
    } else {
        fmt.Printf("Falied to open YAML file: %v\n", fileErr)
    }

    return &config
}
