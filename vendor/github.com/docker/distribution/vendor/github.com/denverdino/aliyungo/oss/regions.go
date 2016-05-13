package oss

import (
	"fmt"
)

// Region represents OSS region
type Region string

// Constants of region definition
const (
	Hangzhou     = Region("oss-cn-hangzhou")
	Qingdao      = Region("oss-cn-qingdao")
	Beijing      = Region("oss-cn-beijing")
	Hongkong     = Region("oss-cn-hongkong")
	Shenzhen     = Region("oss-cn-shenzhen")
	USWest1      = Region("oss-us-west-1")
	APSouthEast1 = Region("oss-ap-southeast-1")
	Shanghai     = Region("oss-cn-shanghai")

	DefaultRegion = Hangzhou
)

// GetEndpoint returns endpoint of region
func (r Region) GetEndpoint(internal bool, bucket string, secure bool) string {
	if internal {
		return r.GetInternalEndpoint(bucket, secure)
	}
	return r.GetInternetEndpoint(bucket, secure)
}

func getProtocol(secure bool) string {
	protocol := "http"
	if secure {
		protocol = "https"
	}
	return protocol
}

// GetInternetEndpoint returns internet endpoint of region
func (r Region) GetInternetEndpoint(bucket string, secure bool) string {
	protocol := getProtocol(secure)
	if bucket == "" {
		return fmt.Sprintf("%s://oss.aliyuncs.com", protocol)
	}
	return fmt.Sprintf("%s://%s.%s.aliyuncs.com", protocol, bucket, string(r))
}

// GetInternalEndpoint returns internal endpoint of region
func (r Region) GetInternalEndpoint(bucket string, secure bool) string {
	protocol := getProtocol(secure)
	if bucket == "" {
		return fmt.Sprintf("%s://oss-internal.aliyuncs.com", protocol)
	}
	return fmt.Sprintf("%s://%s.%s-internal.aliyuncs.com", protocol, bucket, string(r))
}
