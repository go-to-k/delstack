package option

import "runtime"

type Option struct {
	Profile   string `short:"p" long:"profile" description:"AWS profile name"`
	StackName string `short:"s" long:"stackName" description:"Stack name" required:"true"`
	Region    string `short:"r" long:"region" description:"AWS Region" default:"ap-northeast-1"`
}

var CONCURRENCY_NUM = runtime.NumCPU()
