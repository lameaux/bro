package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Lameaux/loadpro/internal/config"
	"github.com/Lameaux/loadpro/internal/runner"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	version = "v0.0.1"
)

var GitHash string

func main() {
	fmt.Printf("loadpro %s, build %s\n", version, GitHash)

	var configFile string
	flag.StringVar(&configFile, "config", "", "Config YAML file")
	flag.Parse()

	if configFile == "" {
		log.Fatal("--config parameter is missing. Example: loadpro --config myconfig.yaml")
	}

	c, err := config.LoadFromFile(configFile)
	if err != nil {
		log.Fatalf("Error loading config from file: %v", err)
	}

	fmt.Printf("Config: %s\n", c.Name)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		fmt.Printf("Received signal: %v\n", sig)
		cancel()
	}()

	fmt.Printf("Executing scenarios... Press Ctrl+C (SIGINT) or send SIGTERM to terminate.\n")
	for _, scenario := range c.Scenarios {
		err = runner.RunScenario(ctx, scenario)
		if err != nil {
			log.Fatalf("Failed to run scenario: %v", err)
		}
	}

}
