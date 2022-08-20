package main

import (
	"os"

	"github.com/go-to-k/delstack/operations"
	"github.com/go-to-k/delstack/option"
	flags "github.com/jessevdk/go-flags"
)

/*
	TODO: add logger
*/

func main() {
	var opts option.Option
	_, err := flags.Parse(&opts)
	if err != nil {
		// os.Exit(1)
		return
	}

	config, err := operations.LoadAwsConfig(opts.Profile, opts.Region)
	if err != nil {
		os.Exit(1)
	}

	if err := operations.DeleteStackResources(config, opts.StackName); err != nil {
		os.Exit(1)
	}
}
