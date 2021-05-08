package adapter

import (
    "fmt"
    "gopkg.in/yaml.v3"
    "io/ioutil"
    "os"
    "strconv"
    "time"
)

type Config struct {
    Adapter AdapterConfig `yaml:"adapter"`
}

type AdapterConfig struct {
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
    ac := AdapterConfig {
        AdapterAddress: "0.0.0.0:3333",
        SulAddress:     "implementation:4433",
        SulName:        "quic.tiferrei.com",
        HTTP3:          false,
        HttpPath:       "/index.html",
        Tracing:        false,
        WaitTime:       waitTime,
    }

    return Config{ac}
}

func (ac* AdapterConfig) Print() {
    fmt.Printf("Adapter Config:\n")
    fmt.Printf("Adapter Address: %v\n", ac.AdapterAddress)
    fmt.Printf("Sul Address: %v\n", ac.SulAddress)
    fmt.Printf("Sul Name: %v\n", ac.SulName)
    fmt.Printf("HTTP3: %v\n", ac.HTTP3)
    fmt.Printf("Http Path: %v\n", ac.HttpPath)
    fmt.Printf("Tracing: %v\n", ac.Tracing)
    fmt.Printf("waitTime: %v\n", ac.WaitTime.String())
}

func GetConfig(path string) *AdapterConfig {
    config := newConfig()

    yamlFile, fileErr := ioutil.ReadFile(path)
    if fileErr == nil {
        yamlErr := yaml.Unmarshal(yamlFile, config)
        if yamlErr != nil {
            // reset config on failure.
            config = newConfig()
        }
    }

    adapterAddress, exists := os.LookupEnv("ADAPTER_ADDRESS")
    if exists {
        config.Adapter.AdapterAddress = adapterAddress
    }

    sulName, exists := os.LookupEnv("SUL_NAME")
    if exists {
        config.Adapter.SulName = sulName
    }

    http3, exists := os.LookupEnv("HTTP3")
    if exists {
        http3Bool, err := strconv.ParseBool(http3)
        if err == nil {
            config.Adapter.HTTP3 = http3Bool
        }
    }

    httpPath, exists := os.LookupEnv("HTTP_PATH")
    if exists {
        config.Adapter.HttpPath = httpPath
    }

    tracing, exists := os.LookupEnv("TRACING")
    if exists {
        tracingBool, err := strconv.ParseBool(tracing)
        if err == nil {
            config.Adapter.Tracing = tracingBool
        }
    }

    waitTime, exists := os.LookupEnv("WAIT_TIME")
    if exists {
        duration, err := time.ParseDuration(waitTime)
        if err == nil {
            config.Adapter.WaitTime = duration
        }
    }

    return &config.Adapter
}
