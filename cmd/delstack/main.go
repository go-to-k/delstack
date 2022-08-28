package main

import (
	"log"
	"os"

	"github.com/go-to-k/delstack/operation"
	"github.com/go-to-k/delstack/option"
	flags "github.com/jessevdk/go-flags"
)

/*
	TODO: add logger
	TODO: message handler or high usability messages
*/

func main() {
	var opts option.Option
	_, err := flags.Parse(&opts)
	if err != nil {
		// os.Exit(1)
		// TODO: show help message
		return
	}

	config, err := operation.LoadAwsConfig(opts.Profile, opts.Region)
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
		os.Exit(1)
	}

	if err := operation.DeleteStackResources(config, opts.StackName); err != nil {
		log.Fatalf("Error: %s", err.Error())
		os.Exit(1)
	}
}
