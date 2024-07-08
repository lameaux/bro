package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Lameaux/bro/internal/config"
	"github.com/Lameaux/bro/internal/runner"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	logo = `
 _               
 | |              
 | |__  _ __ ___  
 | '_ \| '__/ _ \ 
 | |_) | | | (_) |
 |_.__/|_|  \___/ 
`
	app     = "bro"
	version = "v0.0.1"
)

var GitHash string

func main() {
	fmt.Printf("%s %s %s, build %s\n", logo, app, version, GitHash)

	var configFile string
	flag.StringVar(&configFile, "config", "", "Config YAML file")
	flag.Parse()

	if configFile == "" {
		log.Fatalf("--config parameter is missing. Example: %s --config myconfig.yaml", app)
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
