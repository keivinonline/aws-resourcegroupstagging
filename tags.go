package main

import (
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
)

// tagExists looks for tagKey between tags
func tagExists(tags []*resourcegroupstaggingapi.Tag, tagKey *string) bool {
	for _, tag := range tags {
		if *tag.Key == *tagKey {
			return true
		}
	}
	return false
}

// create a function that checks for tags and value
// and returns a bool
func tagValueExists(tags []*resourcegroupstaggingapi.Tag, tagKey *string, tagValue *string) bool {
	for _, tag := range tags {
		if (*tag.Key == *tagKey) && (*tag.Value == *tagValue) {
			return true
		}
	}
	return false
}
