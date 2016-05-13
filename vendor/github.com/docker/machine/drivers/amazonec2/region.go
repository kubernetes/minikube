package amazonec2

import (
	"errors"
)

type region struct {
	AmiId string
}

// Release 15.10 20151116.1
// See https://cloud-images.ubuntu.com/locator/ec2/
var regionDetails map[string]*region = map[string]*region{
	"ap-northeast-1": {"ami-b36d4edd"},
	"ap-northeast-2": {"ami-09dc1267"},
	"ap-southeast-1": {"ami-1069af73"},
	"ap-southeast-2": {"ami-1d336a7e"},
	"cn-north-1":     {"ami-79eb2214"},
	"eu-west-1":      {"ami-8aa67cf9"},
	"eu-central-1":   {"ami-ab0210c7"},
	"sa-east-1":      {"ami-185de774"},
	"us-east-1":      {"ami-26d5af4c"},
	"us-west-1":      {"ami-9cbcd2fc"},
	"us-west-2":      {"ami-16b1a077"},
	"us-gov-west-1":  {"ami-b0bad893"},
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
