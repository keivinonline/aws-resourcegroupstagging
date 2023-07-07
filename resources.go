package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
)

// printResource prints a resource
func printResource(resource *resourcegroupstaggingapi.ResourceTagMapping) {
	fmt.Printf("%s\n", *resource.ResourceARN)
}

func getResourceARN(resource *resourcegroupstaggingapi.ResourceTagMapping) string {
	return *resource.ResourceARN
}
