package amazonec2

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/stretchr/testify/mock"
)

type fakeEC2 struct {
	*ec2.EC2
}

type errorProvider struct{}

func (p *errorProvider) Retrieve() (credentials.Value, error) {
	return credentials.Value{}, errors.New("bad credentials")
}

func (p *errorProvider) IsExpired() bool {
	return true
}

type okProvider struct {
	accessKeyID     string
	secretAccessKey string
	sessionToken    string
}

func (p *okProvider) Retrieve() (credentials.Value, error) {
	return credentials.Value{
		AccessKeyID:     p.accessKeyID,
		SecretAccessKey: p.secretAccessKey,
		SessionToken:    p.sessionToken,
	}, nil
}

func (p *okProvider) IsExpired() bool {
	return true
}

type cliCredentials struct{}

func (c *cliCredentials) NewStaticCredentials(id, secret, token string) *credentials.Credentials {
	return credentials.NewCredentials(&okProvider{id, secret, token})
}

func (c *cliCredentials) NewSharedCredentials(filename, profile string) *credentials.Credentials {
	return credentials.NewCredentials(&errorProvider{})
}

type fileCredentials struct{}

func (c *fileCredentials) NewStaticCredentials(id, secret, token string) *credentials.Credentials {
	return nil
}

func (c *fileCredentials) NewSharedCredentials(filename, profile string) *credentials.Credentials {
	return credentials.NewCredentials(&okProvider{"access", "secret", "token"})
}

type fakeEC2WithDescribe struct {
	*fakeEC2
	output *ec2.DescribeAccountAttributesOutput
	err    error
}

func (f *fakeEC2WithDescribe) DescribeAccountAttributes(input *ec2.DescribeAccountAttributesInput) (*ec2.DescribeAccountAttributesOutput, error) {
	return f.output, f.err
}

type fakeEC2WithLogin struct {
	*fakeEC2
}

func (f *fakeEC2WithLogin) DescribeAccountAttributes(input *ec2.DescribeAccountAttributesInput) (*ec2.DescribeAccountAttributesOutput, error) {
	defaultVpc := "default-vpc"
	vpcName := "vpc-9999"

	return &ec2.DescribeAccountAttributesOutput{
		AccountAttributes: []*ec2.AccountAttribute{
			{
				AttributeName: &defaultVpc,
				AttributeValues: []*ec2.AccountAttributeValue{
					{AttributeValue: &vpcName},
				},
			},
		},
	}, nil
}

type fakeEC2SecurityGroupTestRecorder struct {
	*fakeEC2
	mock.Mock
}

func (f *fakeEC2SecurityGroupTestRecorder) DescribeSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	result := f.Called(input)
	err := result.Error(1)
	value, ok := result.Get(0).(*ec2.DescribeSecurityGroupsOutput)
	if !ok && err == nil {
		return nil, errors.New("Type assertion to DescribeSecurityGroupsOutput failed")
	}
	return value, err
}

func (f *fakeEC2SecurityGroupTestRecorder) CreateSecurityGroup(input *ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error) {
	result := f.Called(input)
	err := result.Error(1)
	value, ok := result.Get(0).(*ec2.CreateSecurityGroupOutput)
	if !ok && err == nil {
		return nil, errors.New("Type assertion to CreateSecurityGroupOutput failed")
	}
	return value, err
}

func (f *fakeEC2SecurityGroupTestRecorder) AuthorizeSecurityGroupIngress(input *ec2.AuthorizeSecurityGroupIngressInput) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	result := f.Called(input)
	err := result.Error(1)
	value, ok := result.Get(0).(*ec2.AuthorizeSecurityGroupIngressOutput)
	if !ok && err == nil {
		return nil, errors.New("Type assertion to AuthorizeSecurityGroupIngressInput failed")
	}
	return value, err
}

func NewTestDriver() *Driver {
	driver := NewDriver("machineFoo", "path")
	driver.clientFactory = func() Ec2Client { return &fakeEC2{} }
	return driver
}

func NewCustomTestDriver(ec2Client Ec2Client) *Driver {
	driver := NewDriver("machineFoo", "path")
	driver.clientFactory = func() Ec2Client { return ec2Client }
	return driver
}
