package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/joho/godotenv"
)

func init() {
	// Load godotenv
	err := godotenv.Load(
		".env",
		".tags.env",
	)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	var err error

	// fetch tags and convert to map
	tagKeysList := strings.Split(os.Getenv("TAG_KEYS"), ",")
	tagValuesList := strings.Split(os.Getenv("TAG_VALUES"), ",")
	tagMap, ok := combineListsToMap(tagKeysList, &tagValuesList)
	if !ok {
		fmt.Println("Error: tagKeysList and tagValuesList are not the same length")
	}

	// get the region flag
	regionFlag := flag.String("region", "", "AWS Region code")

	// get the tag key flag
	tagKeyFlag := flag.String("tag", "", "Tag key to find")

	// get the tag key flag
	tagValueFlag := flag.String("value", "", "Tag value to find")

	// get the untagged flag
	untaggedFlag := flag.Bool("untagged", false, "Display resources without tag if set")

	// get the applyTags flag
	applyTags := flag.Bool("applytags", false, "Apply tags to resources if set")

	// get the extra resources flag
	extra := flag.Bool("extra", false, "Additional resources to tag")
	// parse all the flags
	flag.Parse()

	// exit if missing parameters
	if *regionFlag == "" || *tagKeyFlag == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// pagination and Input/Output variables
	var paginationToken string = ""
	var resourcesInput *resourcegroupstaggingapi.GetResourcesInput
	var resourcesOutput *resourcegroupstaggingapi.GetResourcesOutput
	var counter int = 0

	// tagging Input/Output variables
	var availableResourceARNList []*string
	var availableResourceARNListFinal []*string
	var taggedResourceARNList []*string

	failureInfoMap := make(map[string]string)

	// extra resources to tag
	var extraResourcesARNList []*string

	// create the session
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// create the ResourceGroupsTaggingAPI client
	client := resourcegroupstaggingapi.New(session, aws.NewConfig().WithRegion(*regionFlag))

	if *extra {
		ec2Client := ec2.New(session, aws.NewConfig().WithRegion(*regionFlag))
		results, ok := ec2Client.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{})
		if ok != nil {
			fmt.Println("Error: could not describe VPC endpoints")
		}
		// fmt.Println(results.VpcEndpoints)
		for _, v := range results.VpcEndpoints {
			arn := fmt.Sprintf("arn:aws:ec2:%v:%v:vpc-endpoint/%v", *regionFlag, *v.OwnerId, *v.VpcEndpointId)
			extraResourcesARNList = append(extraResourcesARNList, &arn)
		}

	}

	// loop over resources
	for {

		// request input
		resourcesInput = &resourcegroupstaggingapi.GetResourcesInput{
			ResourcesPerPage: aws.Int64(100),
			PaginationToken:  &paginationToken,
		}

		// retrieve all resources
		resourcesOutput, err = client.GetResources(resourcesInput)
		if err != nil {
			fmt.Println(err)
		}

		// case 1) if I want to find a tag value
		// for every resource
		for _, resource := range resourcesOutput.ResourceTagMappingList {

			if *tagKeyFlag == "" {
				// analyze tags
				tagFound := tagExists(resource.Tags, tagKeyFlag)

				// if the resource is tagged and I want untagged resource then skip
				if tagFound && *untaggedFlag {
					continue
				}

				// if the resource is not tagged and I want tagged resource then skip
				if !tagFound && !*untaggedFlag {
					continue
				}

				// printResource(resource) // to be removed ?
				counter = counter + 1
				// add resource to list
				availableResourceARNList = append(availableResourceARNList, resource.ResourceARN)
			} else {
				// analyze tags values
				// analyze tags
				tagAndValueFound := tagValueExists(resource.Tags, tagKeyFlag, tagValueFlag)

				// if the resource is tagged and I want untagged resource then skip
				if tagAndValueFound && *untaggedFlag {
					continue
				}

				// if the resource is not tagged and I want tagged resource then skip
				if !tagAndValueFound && !*untaggedFlag {
					continue
				}

				// printResource(resource) // to be removed ?
				counter = counter + 1
				// add resource to list
				availableResourceARNList = append(availableResourceARNList, resource.ResourceARN)

			}
		}
		// case 2) if I want to find a tag key and value
		// for _, resource := range resourcesOutput.ResourceTagMappingList {

		// 	// analyze tags
		// 	tagAndValueFound := tagValueExists(resource.Tags, tagKeyFlag, tagValueFlag)

		// 	// if the resource is tagged and I want untagged resource then skip
		// 	if tagAndValueFound && *untaggedFlag {
		// 		continue
		// 	}

		// 	// if the resource is not tagged and I want tagged resource then skip
		// 	if !tagAndValueFound && !*untaggedFlag {
		// 		continue
		// 	}

		// 	// printResource(resource) // to be removed ?
		// 	counter = counter + 1
		// 	// add resource to list
		// 	availableResourceARNList = append(availableResourceARNList, resource.ResourceARN)
		// }
		// loop until the paginationToken is empty, no more pages
		paginationToken = *resourcesOutput.PaginationToken
		if *resourcesOutput.PaginationToken == "" {
			break
		}
	}

	// combine list
	availableResourceARNListFinal = append(availableResourceARNList, extraResourcesARNList...)
	// Apply tags section
	if *applyTags {
		for i := 0; i < len(availableResourceARNListFinal); i += 20 {
			start := i
			end := i + 20
			if end > len(availableResourceARNListFinal) {
				end = len(availableResourceARNListFinal)
			}
			fmt.Printf("start: %v, end: %v\n", start, end)
			ARNListToProcess := availableResourceARNListFinal[start:end]

			// create a resourcegroupstaggingapi.TagResourcesInput object
			// and pass it to the TagResources method
			customTagResourcesInput := &resourcegroupstaggingapi.TagResourcesInput{
				ResourceARNList: ARNListToProcess,
				// ResourceARNList: []*string{
				// 	// aws.String(awsResource),
				// 	resourceArn,
				// },
				Tags: tagMap,
			}
			k, err := client.TagResources(customTagResourcesInput)
			if err != nil {
				fmt.Println(err)
			}

			// check len of k.FailedResourcesMap
			if len(k.FailedResourcesMap) != 0 {
				// fmt.Println("TagResources failed")
				// loop and check the output from TagResources
				for arn, v := range k.FailedResourcesMap {
					failureInfoMap[arn] = v.String()
				}
			}
		}
	}
	// }
	fmt.Printf("\n## Tags ##\n")
	// fmt.Println(tagMap)
	for k, v := range tagMap {
		fmt.Printf("%s: %s\n", k, *v)
	}
	fmt.Println()
	fmt.Printf("\n## Summary ##\n")
	fmt.Printf("Total resources mactched via resource tagging API: %v\n", len(availableResourceARNList))
	fmt.Printf("Extra resources matched: %v\n", len(extraResourcesARNList))
	fmt.Printf("Total sum : %v\n", len(availableResourceARNListFinal))

	fmt.Printf(`Looking for tag "%s" between resources in "%s"... `, *tagKeyFlag, *regionFlag)
	if *untaggedFlag {
		fmt.Printf("found %d untagged resources!\n", counter)
	} else {
		fmt.Printf("found %d tagged resources!\n", counter)
	}

	fmt.Printf("Failed resource map items: %v\n", len(failureInfoMap))
	fmt.Printf("\n## Available Resources ARN ##\n")
	for _, value := range availableResourceARNList {
		fmt.Println(*value)
	}
	fmt.Printf("\n## Extra Resources ARN ##\n")
	for _, value := range extraResourcesARNList {
		fmt.Println(*value)
	}

	fmt.Printf("\n## Final Available Resources ARN ##\n")
	for _, value := range availableResourceARNListFinal {
		fmt.Println(*value)
	}

	fmt.Printf("\n## Tagged Resources ARN ##\n")
	for _, value := range taggedResourceARNList {
		fmt.Println(*value)
	}

	fmt.Printf("\n## Tagging errors ##\n")
	for k, v := range failureInfoMap {
		fmt.Printf("ARN: %s\n", k)
		fmt.Printf("Error: %s\n", v)
	}

}
