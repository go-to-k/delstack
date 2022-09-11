package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/operation"
	"github.com/go-to-k/delstack/option"
)

/*
	TODO: add logger
	TODO: message handler or high usability messages
	TODO: os.Exit(1) or not when exit 1
*/

func main() {
	opts := option.NewOption()
	_, err := opts.Parse()
	if err != nil {
		// os.Exit(1)
		return
	}

	config, err := opts.LoadAwsConfig()
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
		os.Exit(1)
	}

	cfnOperator := operation.NewStackOperator(config)
	isRootStack := true
	if err := cfnOperator.DeleteStackResources(aws.String(opts.StackName), isRootStack); err != nil {
		log.Fatalf("Error: %s", err.Error())
		os.Exit(1)
	}
}
