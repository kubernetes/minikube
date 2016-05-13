package amazonec2

import "github.com/aws/aws-sdk-go/service/ec2"

type Ec2Client interface {
	DescribeAccountAttributes(input *ec2.DescribeAccountAttributesInput) (*ec2.DescribeAccountAttributesOutput, error)

	DescribeSubnets(input *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error)

	CreateTags(input *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error)

	//SecurityGroup

	CreateSecurityGroup(input *ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error)

	AuthorizeSecurityGroupIngress(input *ec2.AuthorizeSecurityGroupIngressInput) (*ec2.AuthorizeSecurityGroupIngressOutput, error)

	DescribeSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error)

	DeleteSecurityGroup(input *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error)

	//KeyPair

	DeleteKeyPair(input *ec2.DeleteKeyPairInput) (*ec2.DeleteKeyPairOutput, error)

	ImportKeyPair(input *ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error)

	DescribeKeyPairs(input *ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error)

	//Instances

	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)

	StartInstances(input *ec2.StartInstancesInput) (*ec2.StartInstancesOutput, error)

	RebootInstances(input *ec2.RebootInstancesInput) (*ec2.RebootInstancesOutput, error)

	StopInstances(input *ec2.StopInstancesInput) (*ec2.StopInstancesOutput, error)

	RunInstances(input *ec2.RunInstancesInput) (*ec2.Reservation, error)

	TerminateInstances(input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error)

	//SpotInstances

	RequestSpotInstances(input *ec2.RequestSpotInstancesInput) (*ec2.RequestSpotInstancesOutput, error)

	DescribeSpotInstanceRequests(input *ec2.DescribeSpotInstanceRequestsInput) (*ec2.DescribeSpotInstanceRequestsOutput, error)

	WaitUntilSpotInstanceRequestFulfilled(input *ec2.DescribeSpotInstanceRequestsInput) error
}
