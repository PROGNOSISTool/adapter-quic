package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/PROGNOSISTool/adapter-quic/adapter"
)

func main() {
    config := adapter.GetConfig("config.yaml")

    sulAdapter, err := adapter.NewAdapter(
        config.AdapterAddress,
        config.SulAddress,
        config.SulName,
        config.HTTP3,
        config.HttpPath,
        config.Tracing,
        config.WaitTime)
    if err != nil {
        fmt.Printf("Failed to create Adapter: %v", err.Error())
    }

	SetupCloseHandler(sulAdapter)
	defer func() {
		if err := recover(); err != nil {
		    sulAdapter.Logger.Printf("Panic detected: %v", err)
			sulAdapter.Stop()
			os.Exit(1)
		}
	}()

	sulAdapter.Run()
}

func SetupCloseHandler(adapter *adapter.Adapter) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		adapter.Stop()
		os.Exit(0)
	}()
}
