package common

// Region represents ECS region
type Region string

// Constants of region definition
const (
	Hangzhou     = Region("cn-hangzhou")
	Qingdao      = Region("cn-qingdao")
	Beijing      = Region("cn-beijing")
	Hongkong     = Region("cn-hongkong")
	Shenzhen     = Region("cn-shenzhen")
	USWest1      = Region("us-west-1")
	APSouthEast1 = Region("ap-southeast-1")
	Shanghai     = Region("cn-shanghai")
)

var ValidRegions = []Region{Hangzhou, Qingdao, Beijing, Shenzhen, Hongkong, Shanghai, USWest1, APSouthEast1}
