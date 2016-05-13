package amazonec2

import (
	"testing"

	"errors"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/docker/machine/commands/commandstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testSSHPort    = int64(22)
	testDockerPort = int64(2376)
	testSwarmPort  = int64(3376)
)

var (
	securityGroup = &ec2.SecurityGroup{
		GroupName: aws.String("test-group"),
		GroupId:   aws.String("12345"),
		VpcId:     aws.String("12345"),
	}
)

func TestConfigureSecurityGroupPermissionsEmpty(t *testing.T) {
	driver := NewTestDriver()

	perms := driver.configureSecurityGroupPermissions(securityGroup)

	assert.Len(t, perms, 2)
}

func TestConfigureSecurityGroupPermissionsSshOnly(t *testing.T) {
	driver := NewTestDriver()
	group := securityGroup
	group.IpPermissions = []*ec2.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(int64(testSSHPort)),
			ToPort:     aws.Int64(int64(testSSHPort)),
		},
	}

	perms := driver.configureSecurityGroupPermissions(group)

	assert.Len(t, perms, 1)
	assert.Equal(t, testDockerPort, *perms[0].FromPort)
}

func TestConfigureSecurityGroupPermissionsDockerOnly(t *testing.T) {
	driver := NewTestDriver()
	group := securityGroup
	group.IpPermissions = []*ec2.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64((testDockerPort)),
			ToPort:     aws.Int64((testDockerPort)),
		},
	}

	perms := driver.configureSecurityGroupPermissions(group)

	assert.Len(t, perms, 1)
	assert.Equal(t, testSSHPort, *perms[0].FromPort)
}

func TestConfigureSecurityGroupPermissionsDockerAndSsh(t *testing.T) {
	driver := NewTestDriver()
	group := securityGroup
	group.IpPermissions = []*ec2.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(testSSHPort),
			ToPort:     aws.Int64(testSSHPort),
		},
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(testDockerPort),
			ToPort:     aws.Int64(testDockerPort),
		},
	}

	perms := driver.configureSecurityGroupPermissions(group)

	assert.Empty(t, perms)
}

func TestConfigureSecurityGroupPermissionsWithSwarm(t *testing.T) {
	driver := NewTestDriver()
	driver.SwarmMaster = true
	group := securityGroup
	group.IpPermissions = []*ec2.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(testSSHPort),
			ToPort:     aws.Int64(testSSHPort),
		},
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(testDockerPort),
			ToPort:     aws.Int64(testDockerPort),
		},
	}

	perms := driver.configureSecurityGroupPermissions(group)

	assert.Len(t, perms, 1)
	assert.Equal(t, testSwarmPort, *perms[0].FromPort)
}

func TestValidateAwsRegionValid(t *testing.T) {
	regions := []string{"eu-west-1", "eu-central-1"}

	for _, region := range regions {
		validatedRegion, err := validateAwsRegion(region)

		assert.NoError(t, err)
		assert.Equal(t, region, validatedRegion)
	}
}

func TestValidateAwsRegionInvalid(t *testing.T) {
	regions := []string{"eu-west-2", "eu-central-2"}

	for _, region := range regions {
		_, err := validateAwsRegion(region)

		assert.EqualError(t, err, "Invalid region specified")
	}
}

func TestFindDefaultVPC(t *testing.T) {
	driver := NewDriver("machineFoo", "path")
	driver.clientFactory = func() Ec2Client { return &fakeEC2WithLogin{} }

	vpc, err := driver.getDefaultVPCId()

	assert.Equal(t, "vpc-9999", vpc)
	assert.NoError(t, err)
}

func TestDefaultVPCIsMissing(t *testing.T) {
	driver := NewDriver("machineFoo", "path")
	driver.clientFactory = func() Ec2Client {
		return &fakeEC2WithDescribe{
			output: &ec2.DescribeAccountAttributesOutput{
				AccountAttributes: []*ec2.AccountAttribute{},
			},
		}
	}

	vpc, err := driver.getDefaultVPCId()

	assert.EqualError(t, err, "No default-vpc attribute")
	assert.Empty(t, vpc)
}

func TestDescribeAccountAttributeFails(t *testing.T) {
	driver := NewDriver("machineFoo", "path")
	driver.clientFactory = func() Ec2Client {
		return &fakeEC2WithDescribe{
			err: errors.New("Not Found"),
		}
	}

	vpc, err := driver.getDefaultVPCId()

	assert.EqualError(t, err, "Not Found")
	assert.Empty(t, vpc)
}

func TestAccessKeyIsMandatory(t *testing.T) {
	driver := NewTestDriver()
	driver.awsCredentials = &cliCredentials{}
	options := &commandstest.FakeFlagger{
		Data: map[string]interface{}{
			"name":             "test",
			"amazonec2-region": "us-east-1",
			"amazonec2-zone":   "e",
		},
	}

	err := driver.SetConfigFromFlags(options)

	assert.Equal(t, err, errorMissingAccessKeyOption)
}

func TestAccessKeyIsMandatoryEvenIfSecretKeyIsPassed(t *testing.T) {
	driver := NewTestDriver()
	driver.awsCredentials = &cliCredentials{}
	options := &commandstest.FakeFlagger{
		Data: map[string]interface{}{
			"name":                 "test",
			"amazonec2-secret-key": "123",
			"amazonec2-region":     "us-east-1",
			"amazonec2-zone":       "e",
		},
	}

	err := driver.SetConfigFromFlags(options)

	assert.Equal(t, err, errorMissingAccessKeyOption)
}

func TestSecretKeyIsMandatory(t *testing.T) {
	driver := NewTestDriver()
	driver.awsCredentials = &cliCredentials{}
	options := &commandstest.FakeFlagger{
		Data: map[string]interface{}{
			"name":                 "test",
			"amazonec2-access-key": "foobar",
			"amazonec2-region":     "us-east-1",
			"amazonec2-zone":       "e",
		},
	}

	err := driver.SetConfigFromFlags(options)

	assert.Equal(t, err, errorMissingSecretKeyOption)
}

func TestLoadingFromCredentialsWorked(t *testing.T) {
	driver := NewCustomTestDriver(&fakeEC2WithLogin{})
	driver.awsCredentials = &fileCredentials{}
	options := &commandstest.FakeFlagger{
		Data: map[string]interface{}{
			"name":             "test",
			"amazonec2-region": "us-east-1",
			"amazonec2-zone":   "e",
		},
	}

	err := driver.SetConfigFromFlags(options)

	assert.NoError(t, err)
	assert.Equal(t, "access", driver.AccessKey)
	assert.Equal(t, "secret", driver.SecretKey)
	assert.Equal(t, "token", driver.SessionToken)
}

func TestPassingBothCLIArgWorked(t *testing.T) {
	driver := NewCustomTestDriver(&fakeEC2WithLogin{})
	driver.awsCredentials = &cliCredentials{}
	options := &commandstest.FakeFlagger{
		Data: map[string]interface{}{
			"name":                 "test",
			"amazonec2-access-key": "foobar",
			"amazonec2-secret-key": "123",
			"amazonec2-region":     "us-east-1",
			"amazonec2-zone":       "e",
		},
	}

	err := driver.SetConfigFromFlags(options)

	assert.NoError(t, err)
	assert.Equal(t, "foobar", driver.AccessKey)
	assert.Equal(t, "123", driver.SecretKey)
}

var values = []string{
	"bob",
	"jake",
	"jill",
}

var pointerSliceTests = []struct {
	input    []string
	expected []*string
}{
	{[]string{}, []*string{}},
	{[]string{values[1]}, []*string{&values[1]}},
	{[]string{values[0], values[2], values[2]}, []*string{&values[0], &values[2], &values[2]}},
}

func TestMakePointerSlice(t *testing.T) {
	for _, tt := range pointerSliceTests {
		actual := makePointerSlice(tt.input)
		assert.Equal(t, tt.expected, actual)
	}
}

var securityGroupNameTests = []struct {
	groupName  string
	groupNames []string
	expected   []string
}{
	{groupName: "bob", expected: []string{"bob"}},
	{groupNames: []string{"bill"}, expected: []string{"bill"}},
	{groupName: "bob", groupNames: []string{"bill"}, expected: []string{"bob", "bill"}},
}

func TestMergeSecurityGroupName(t *testing.T) {
	for _, tt := range securityGroupNameTests {
		d := Driver{SecurityGroupName: tt.groupName, SecurityGroupNames: tt.groupNames}
		assert.Equal(t, tt.expected, d.securityGroupNames())
	}
}

var securityGroupIdTests = []struct {
	groupId  string
	groupIds []string
	expected []string
}{
	{groupId: "id", expected: []string{"id"}},
	{groupIds: []string{"id"}, expected: []string{"id"}},
	{groupId: "id1", groupIds: []string{"id2"}, expected: []string{"id1", "id2"}},
}

func TestMergeSecurityGroupId(t *testing.T) {
	for _, tt := range securityGroupIdTests {
		d := Driver{SecurityGroupId: tt.groupId, SecurityGroupIds: tt.groupIds}
		assert.Equal(t, tt.expected, d.securityGroupIds())
	}
}

func matchGroupLookup(expected []string) interface{} {
	return func(input *ec2.DescribeSecurityGroupsInput) bool {
		actual := []string{}
		for _, filter := range input.Filters {
			if *filter.Name == "group-name" {
				for _, groupName := range filter.Values {
					actual = append(actual, *groupName)
				}
			}
		}
		return reflect.DeepEqual(expected, actual)
	}
}

func ipPermission(port int64) *ec2.IpPermission {
	return &ec2.IpPermission{
		FromPort:   aws.Int64(port),
		ToPort:     aws.Int64(port),
		IpProtocol: aws.String("tcp"),
		IpRanges:   []*ec2.IpRange{{CidrIp: aws.String(ipRange)}},
	}
}

func TestConfigureSecurityGroupsEmpty(t *testing.T) {
	recorder := fakeEC2SecurityGroupTestRecorder{}

	driver := NewCustomTestDriver(&recorder)
	err := driver.configureSecurityGroups([]string{})

	assert.Nil(t, err)
	recorder.AssertExpectations(t)
}

func TestConfigureSecurityGroupsMixed(t *testing.T) {
	groups := []string{"existingGroup", "newGroup"}
	recorder := fakeEC2SecurityGroupTestRecorder{}

	// First, a check is made for which groups already exist.
	initialLookupResult := ec2.DescribeSecurityGroupsOutput{SecurityGroups: []*ec2.SecurityGroup{
		{
			GroupName:     aws.String("existingGroup"),
			GroupId:       aws.String("existingGroupId"),
			IpPermissions: []*ec2.IpPermission{ipPermission(testSSHPort)},
		},
	}}
	recorder.On("DescribeSecurityGroups", mock.MatchedBy(matchGroupLookup(groups))).Return(
		&initialLookupResult, nil)

	// An ingress permission is added to the existing group.
	recorder.On("AuthorizeSecurityGroupIngress", &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       aws.String("existingGroupId"),
		IpPermissions: []*ec2.IpPermission{ipPermission(testDockerPort)},
	}).Return(
		&ec2.AuthorizeSecurityGroupIngressOutput{}, nil)

	// The new security group is created.
	recorder.On("CreateSecurityGroup", &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String("newGroup"),
		Description: aws.String("Docker Machine"),
		VpcId:       aws.String(""),
	}).Return(
		&ec2.CreateSecurityGroupOutput{GroupId: aws.String("newGroupId")}, nil)

	// Ensuring the new security group exists.
	postCreateLookupResult := ec2.DescribeSecurityGroupsOutput{SecurityGroups: []*ec2.SecurityGroup{
		{
			GroupName: aws.String("newGroup"),
			GroupId:   aws.String("newGroupId"),
		},
	}}
	recorder.On("DescribeSecurityGroups",
		&ec2.DescribeSecurityGroupsInput{GroupIds: []*string{aws.String("newGroupId")}}).Return(
		&postCreateLookupResult, nil)

	// Permissions are added to the new security group.
	recorder.On("AuthorizeSecurityGroupIngress", &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       aws.String("newGroupId"),
		IpPermissions: []*ec2.IpPermission{ipPermission(testSSHPort), ipPermission(testDockerPort)},
	}).Return(
		&ec2.AuthorizeSecurityGroupIngressOutput{}, nil)

	driver := NewCustomTestDriver(&recorder)
	err := driver.configureSecurityGroups(groups)

	assert.Nil(t, err)
	recorder.AssertExpectations(t)
}

func TestConfigureSecurityGroupsErrLookupExist(t *testing.T) {
	groups := []string{"group"}
	recorder := fakeEC2SecurityGroupTestRecorder{}

	lookupExistErr := errors.New("lookup failed")
	recorder.On("DescribeSecurityGroups", mock.MatchedBy(matchGroupLookup(groups))).Return(
		nil, lookupExistErr)

	driver := NewCustomTestDriver(&recorder)
	err := driver.configureSecurityGroups(groups)

	assert.Exactly(t, lookupExistErr, err)
	recorder.AssertExpectations(t)
}
