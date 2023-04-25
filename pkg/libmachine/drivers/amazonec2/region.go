package amazonec2

import (
	"errors"
)

type region struct {
	AmiId string
}

// Ubuntu 16.04 LTS 20180228.1 hvm:ebs-ssd (amd64)
// See https://cloud-images.ubuntu.com/locator/ec2/
var regionDetails map[string]*region = map[string]*region{
	"ap-northeast-1":  {"ami-bcb7f6da"},
	"ap-northeast-2":  {"ami-5073de3e"},
	"ap-southeast-1":  {"ami-41e4af3d"},
	"ap-southeast-2":  {"ami-c1498fa3"},
	"ap-south-1":      {"ami-1083dc7f"},
	"ca-central-1":    {"ami-8d9e19e9"},
	"cn-north-1":      {"ami-cc4499a1"}, // Note: this is 20180126
	"cn-northwest-1":  {"ami-fd0e1a9f"}, // Note: this is 20180126
	"eu-north-1":      {"ami-017ff17f"},
	"eu-central-1":    {"ami-bc4925d3"},
	"eu-west-1":       {"ami-0b541372"},
	"eu-west-2":       {"ami-ff46a298"},
	"eu-west-3":       {"ami-9465d3e9"},
	"sa-east-1":       {"ami-b5501bd9"},
	"us-east-1":       {"ami-927185ef"},
	"us-east-2":       {"ami-b9daeddc"},
	"us-west-1":       {"ami-264c4646"},
	"us-west-2":       {"ami-78a22900"},
	"us-gov-west-1":   {"ami-2561ea44"},
	"custom-endpoint": {""},
}

func awsRegionsList() []string {
	var list []string

	for k := range regionDetails {
		list = append(list, k)
	}

	return list
}

func validateAwsRegion(region string) (string, error) {
	for _, v := range awsRegionsList() {
		if v == region {
			return region, nil
		}
	}

	return "", errors.New("Invalid region specified")
}
