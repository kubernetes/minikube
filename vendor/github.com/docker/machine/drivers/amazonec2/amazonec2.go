package amazonec2

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

const (
	driverName               = "amazonec2"
	ipRange                  = "0.0.0.0/0"
	machineSecurityGroupName = "docker-machine"
	defaultAmiId             = "ami-615cb725"
	defaultRegion            = "us-east-1"
	defaultInstanceType      = "t2.micro"
	defaultDeviceName        = "/dev/sda1"
	defaultRootSize          = 16
	defaultVolumeType        = "gp2"
	defaultZone              = "a"
	defaultSecurityGroup     = machineSecurityGroupName
	defaultSSHUser           = "ubuntu"
	defaultSpotPrice         = "0.50"
)

const (
	keypairNotFoundCode = "InvalidKeyPair.NotFound"
)

var (
	dockerPort                  = 2376
	swarmPort                   = 3376
	errorMissingAccessKeyOption = errors.New("amazonec2 driver requires the --amazonec2-access-key option or proper credentials in ~/.aws/credentials")
	errorMissingSecretKeyOption = errors.New("amazonec2 driver requires the --amazonec2-secret-key option or proper credentials in ~/.aws/credentials")
	errorNoVPCIdFound           = errors.New("amazonec2 driver requires either the --amazonec2-subnet-id or --amazonec2-vpc-id option or an AWS Account with a default vpc-id")
)

type Driver struct {
	*drivers.BaseDriver
	clientFactory    func() Ec2Client
	awsCredentials   awsCredentials
	Id               string
	AccessKey        string
	SecretKey        string
	SessionToken     string
	Region           string
	AMI              string
	SSHKeyID         int
	KeyName          string
	InstanceId       string
	InstanceType     string
	PrivateIPAddress string

	// NB: SecurityGroupId expanded from single value to slice on 26 Feb 2016 - we maintain both for host storage backwards compatibility.
	SecurityGroupId  string
	SecurityGroupIds []string

	// NB: SecurityGroupName expanded from single value to slice on 26 Feb 2016 - we maintain both for host storage backwards compatibility.
	SecurityGroupName  string
	SecurityGroupNames []string

	Tags                    string
	ReservationId           string
	DeviceName              string
	RootSize                int64
	VolumeType              string
	IamInstanceProfile      string
	VpcId                   string
	SubnetId                string
	Zone                    string
	keyPath                 string
	RequestSpotInstance     bool
	SpotPrice               string
	PrivateIPOnly           bool
	UsePrivateIP            bool
	UseEbsOptimizedInstance bool
	Monitoring              bool
	SSHPrivateKeyPath       string
	RetryCount              int
}

type clientFactory interface {
	build(d *Driver) Ec2Client
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "amazonec2-access-key",
			Usage:  "AWS Access Key",
			EnvVar: "AWS_ACCESS_KEY_ID",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-secret-key",
			Usage:  "AWS Secret Key",
			EnvVar: "AWS_SECRET_ACCESS_KEY",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-session-token",
			Usage:  "AWS Session Token",
			EnvVar: "AWS_SESSION_TOKEN",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-ami",
			Usage:  "AWS machine image",
			EnvVar: "AWS_AMI",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-region",
			Usage:  "AWS region",
			Value:  defaultRegion,
			EnvVar: "AWS_DEFAULT_REGION",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-vpc-id",
			Usage:  "AWS VPC id",
			EnvVar: "AWS_VPC_ID",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-zone",
			Usage:  "AWS zone for instance (i.e. a,b,c,d,e)",
			Value:  defaultZone,
			EnvVar: "AWS_ZONE",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-subnet-id",
			Usage:  "AWS VPC subnet id",
			EnvVar: "AWS_SUBNET_ID",
		},
		mcnflag.StringSliceFlag{
			Name:   "amazonec2-security-group",
			Usage:  "AWS VPC security group",
			Value:  []string{defaultSecurityGroup},
			EnvVar: "AWS_SECURITY_GROUP",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-tags",
			Usage:  "AWS Tags (e.g. key1,value1,key2,value2)",
			EnvVar: "AWS_TAGS",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-instance-type",
			Usage:  "AWS instance type",
			Value:  defaultInstanceType,
			EnvVar: "AWS_INSTANCE_TYPE",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-device-name",
			Usage:  "AWS root device name",
			Value:  defaultDeviceName,
			EnvVar: "AWS_DEVICE_NAME",
		},
		mcnflag.IntFlag{
			Name:   "amazonec2-root-size",
			Usage:  "AWS root disk size (in GB)",
			Value:  defaultRootSize,
			EnvVar: "AWS_ROOT_SIZE",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-volume-type",
			Usage:  "Amazon EBS volume type",
			Value:  defaultVolumeType,
			EnvVar: "AWS_VOLUME_TYPE",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-iam-instance-profile",
			Usage:  "AWS IAM Instance Profile",
			EnvVar: "AWS_INSTANCE_PROFILE",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-ssh-user",
			Usage:  "Set the name of the ssh user",
			Value:  defaultSSHUser,
			EnvVar: "AWS_SSH_USER",
		},
		mcnflag.BoolFlag{
			Name:  "amazonec2-request-spot-instance",
			Usage: "Set this flag to request spot instance",
		},
		mcnflag.StringFlag{
			Name:  "amazonec2-spot-price",
			Usage: "AWS spot instance bid price (in dollar)",
			Value: defaultSpotPrice,
		},
		mcnflag.BoolFlag{
			Name:  "amazonec2-private-address-only",
			Usage: "Only use a private IP address",
		},
		mcnflag.BoolFlag{
			Name:  "amazonec2-use-private-address",
			Usage: "Force the usage of private IP address",
		},
		mcnflag.BoolFlag{
			Name:  "amazonec2-monitoring",
			Usage: "Set this flag to enable CloudWatch monitoring",
		},
		mcnflag.BoolFlag{
			Name:  "amazonec2-use-ebs-optimized-instance",
			Usage: "Create an EBS optimized instance",
		},
		mcnflag.StringFlag{
			Name:   "amazonec2-ssh-keypath",
			Usage:  "SSH Key for Instance",
			EnvVar: "AWS_SSH_KEYPATH",
		},
		mcnflag.IntFlag{
			Name:  "amazonec2-retries",
			Usage: "Set retry count for recoverable failures (use -1 to disable)",
			Value: 5,
		},
	}
}

func NewDriver(hostName, storePath string) *Driver {
	id := generateId()
	driver := &Driver{
		Id:                 id,
		AMI:                defaultAmiId,
		Region:             defaultRegion,
		InstanceType:       defaultInstanceType,
		RootSize:           defaultRootSize,
		Zone:               defaultZone,
		SecurityGroupNames: []string{defaultSecurityGroup},
		SpotPrice:          defaultSpotPrice,
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		},
		awsCredentials: &defaultAWSCredentials{},
	}

	driver.clientFactory = driver.buildClient

	return driver
}

func (d *Driver) buildClient() Ec2Client {
	config := aws.NewConfig()
	alogger := AwsLogger()
	config = config.WithRegion(d.Region)
	config = config.WithCredentials(d.awsCredentials.NewStaticCredentials(d.AccessKey, d.SecretKey, d.SessionToken))
	config = config.WithLogger(alogger)
	config = config.WithLogLevel(aws.LogDebugWithHTTPBody)
	config = config.WithMaxRetries(d.RetryCount)
	return ec2.New(session.New(config))
}

func (d *Driver) getClient() Ec2Client {
	return d.clientFactory()
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	region, err := validateAwsRegion(flags.String("amazonec2-region"))
	if err != nil {
		return err
	}

	image := flags.String("amazonec2-ami")
	if len(image) == 0 {
		image = regionDetails[region].AmiId
	}

	d.AccessKey = flags.String("amazonec2-access-key")
	d.SecretKey = flags.String("amazonec2-secret-key")
	d.SessionToken = flags.String("amazonec2-session-token")
	d.Region = region
	d.AMI = image
	d.RequestSpotInstance = flags.Bool("amazonec2-request-spot-instance")
	d.SpotPrice = flags.String("amazonec2-spot-price")
	d.InstanceType = flags.String("amazonec2-instance-type")
	d.VpcId = flags.String("amazonec2-vpc-id")
	d.SubnetId = flags.String("amazonec2-subnet-id")
	d.SecurityGroupNames = flags.StringSlice("amazonec2-security-group")
	d.Tags = flags.String("amazonec2-tags")
	zone := flags.String("amazonec2-zone")
	d.Zone = zone[:]
	d.DeviceName = flags.String("amazonec2-device-name")
	d.RootSize = int64(flags.Int("amazonec2-root-size"))
	d.VolumeType = flags.String("amazonec2-volume-type")
	d.IamInstanceProfile = flags.String("amazonec2-iam-instance-profile")
	d.SSHUser = flags.String("amazonec2-ssh-user")
	d.SSHPort = 22
	d.PrivateIPOnly = flags.Bool("amazonec2-private-address-only")
	d.UsePrivateIP = flags.Bool("amazonec2-use-private-address")
	d.Monitoring = flags.Bool("amazonec2-monitoring")
	d.UseEbsOptimizedInstance = flags.Bool("amazonec2-use-ebs-optimized-instance")
	d.SSHPrivateKeyPath = flags.String("amazonec2-ssh-keypath")
	d.SetSwarmConfigFromFlags(flags)
	d.RetryCount = flags.Int("amazonec2-retries")

	if d.AccessKey == "" && d.SecretKey == "" {
		credentials, err := d.awsCredentials.NewSharedCredentials("", "").Get()
		if err != nil {
			log.Debug("Could not load credentials from ~/.aws/credentials")
		} else {
			log.Debug("Successfully loaded credentials from ~/.aws/credentials")
			d.AccessKey = credentials.AccessKeyID
			d.SecretKey = credentials.SecretAccessKey
			d.SessionToken = credentials.SessionToken
		}
	}

	if d.AccessKey == "" {
		return errorMissingAccessKeyOption
	}

	if d.SecretKey == "" {
		return errorMissingSecretKeyOption
	}

	if d.VpcId == "" {
		d.VpcId, err = d.getDefaultVPCId()
		if err != nil {
			log.Warnf("Couldn't determine your account Default VPC ID : %q", err)
		}
	}

	if d.SubnetId == "" && d.VpcId == "" {
		return errorNoVPCIdFound
	}

	if d.SubnetId != "" && d.VpcId != "" {
		subnetFilter := []*ec2.Filter{
			{
				Name:   aws.String("subnet-id"),
				Values: []*string{&d.SubnetId},
			},
		}

		subnets, err := d.getClient().DescribeSubnets(&ec2.DescribeSubnetsInput{
			Filters: subnetFilter,
		})
		if err != nil {
			return err
		}

		if *subnets.Subnets[0].VpcId != d.VpcId {
			return fmt.Errorf("SubnetId: %s does not belong to VpcId: %s", d.SubnetId, d.VpcId)
		}
	}

	if d.isSwarmMaster() {
		u, err := url.Parse(d.SwarmHost)
		if err != nil {
			return fmt.Errorf("error parsing swarm host: %s", err)
		}

		parts := strings.Split(u.Host, ":")
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}

		swarmPort = port
	}

	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) checkPrereqs() error {
	// check for existing keypair
	key, err := d.getClient().DescribeKeyPairs(&ec2.DescribeKeyPairsInput{
		KeyNames: []*string{&d.MachineName},
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == keypairNotFoundCode {
			// Not a real error for 'NotFound' since we're checking existance anyways
		} else {
			return err
		}
	}

	if err == nil && len(key.KeyPairs) != 0 {
		return fmt.Errorf("There is already a keypair with the name %s.  Please either remove that keypair or use a different machine name.", d.MachineName)
	}

	regionZone := d.Region + d.Zone
	if d.SubnetId == "" {
		filters := []*ec2.Filter{
			{
				Name:   aws.String("availability-zone"),
				Values: []*string{&regionZone},
			},
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{&d.VpcId},
			},
		}

		subnets, err := d.getClient().DescribeSubnets(&ec2.DescribeSubnetsInput{
			Filters: filters,
		})
		if err != nil {
			return err
		}

		if len(subnets.Subnets) == 0 {
			return fmt.Errorf("unable to find a subnet in the zone: %s", regionZone)
		}

		d.SubnetId = *subnets.Subnets[0].SubnetId

		// try to find default
		if len(subnets.Subnets) > 1 {
			for _, subnet := range subnets.Subnets {
				if *subnet.DefaultForAz {
					d.SubnetId = *subnet.SubnetId
					break
				}
			}
		}
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	return d.checkPrereqs()
}

func (d *Driver) instanceIpAvailable() bool {
	ip, err := d.GetIP()
	if err != nil {
		log.Debug(err)
	}
	if ip != "" {
		d.IPAddress = ip
		log.Debugf("Got the IP Address, it's %q", d.IPAddress)
		return true
	}
	return false
}

func makePointerSlice(stackSlice []string) []*string {
	pointerSlice := []*string{}
	for i := range stackSlice {
		pointerSlice = append(pointerSlice, &stackSlice[i])
	}
	return pointerSlice
}

// Support migrating single string Driver fields to slices.
func migrateStringToSlice(value string, values []string) (result []string) {
	if value != "" {
		result = append(result, value)
	}
	result = append(result, values...)
	return
}

func (d *Driver) securityGroupNames() (ids []string) {
	return migrateStringToSlice(d.SecurityGroupName, d.SecurityGroupNames)
}

func (d *Driver) securityGroupIds() (ids []string) {
	return migrateStringToSlice(d.SecurityGroupId, d.SecurityGroupIds)
}

func (d *Driver) Create() error {
	if err := d.checkPrereqs(); err != nil {
		return err
	}

	log.Infof("Launching instance...")

	if err := d.createKeyPair(); err != nil {
		return fmt.Errorf("unable to create key pair: %s", err)
	}

	if err := d.configureSecurityGroups(d.securityGroupNames()); err != nil {
		return err
	}

	bdm := &ec2.BlockDeviceMapping{
		DeviceName: aws.String(d.DeviceName),
		Ebs: &ec2.EbsBlockDevice{
			VolumeSize:          aws.Int64(d.RootSize),
			VolumeType:          aws.String(d.VolumeType),
			DeleteOnTermination: aws.Bool(true),
		},
	}
	netSpecs := []*ec2.InstanceNetworkInterfaceSpecification{{
		DeviceIndex:              aws.Int64(0), // eth0
		Groups:                   makePointerSlice(d.securityGroupIds()),
		SubnetId:                 &d.SubnetId,
		AssociatePublicIpAddress: aws.Bool(!d.PrivateIPOnly),
	}}

	regionZone := d.Region + d.Zone
	log.Debugf("launching instance in subnet %s", d.SubnetId)

	var instance *ec2.Instance

	if d.RequestSpotInstance {
		spotInstanceRequest, err := d.getClient().RequestSpotInstances(&ec2.RequestSpotInstancesInput{
			LaunchSpecification: &ec2.RequestSpotLaunchSpecification{
				ImageId: &d.AMI,
				Placement: &ec2.SpotPlacement{
					AvailabilityZone: &regionZone,
				},
				KeyName:           &d.KeyName,
				InstanceType:      &d.InstanceType,
				NetworkInterfaces: netSpecs,
				Monitoring:        &ec2.RunInstancesMonitoringEnabled{Enabled: aws.Bool(d.Monitoring)},
				IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
					Name: &d.IamInstanceProfile,
				},
				EbsOptimized:        &d.UseEbsOptimizedInstance,
				BlockDeviceMappings: []*ec2.BlockDeviceMapping{bdm},
			},
			InstanceCount: aws.Int64(1),
			SpotPrice:     &d.SpotPrice,
		})
		if err != nil {
			return fmt.Errorf("Error request spot instance: %s", err)
		}

		log.Info("Waiting for spot instance...")
		err = d.getClient().WaitUntilSpotInstanceRequestFulfilled(&ec2.DescribeSpotInstanceRequestsInput{
			SpotInstanceRequestIds: []*string{spotInstanceRequest.SpotInstanceRequests[0].SpotInstanceRequestId},
		})
		if err != nil {
			return fmt.Errorf("Error fulfilling spot request: %v", err)
		}
		log.Info("Created spot instance request %v", *spotInstanceRequest.SpotInstanceRequests[0].SpotInstanceRequestId)
		// resolve instance id
		for i := 0; i < 3; i++ {
			// Even though the waiter succeeded, eventual consistency means we could
			// get a describe output that does not include this information. Try a
			// few times just in case
			var resolvedSpotInstance *ec2.DescribeSpotInstanceRequestsOutput
			resolvedSpotInstance, err = d.getClient().DescribeSpotInstanceRequests(&ec2.DescribeSpotInstanceRequestsInput{
				SpotInstanceRequestIds: []*string{spotInstanceRequest.SpotInstanceRequests[0].SpotInstanceRequestId},
			})
			if err != nil {
				// Unexpected; no need to retry
				return fmt.Errorf("Error describing previously made spot instance request: %v", err)
			}
			maybeInstanceId := resolvedSpotInstance.SpotInstanceRequests[0].InstanceId
			if maybeInstanceId != nil {
				var instances *ec2.DescribeInstancesOutput
				instances, err = d.getClient().DescribeInstances(&ec2.DescribeInstancesInput{
					InstanceIds: []*string{maybeInstanceId},
				})
				if err != nil {
					// Retry if we get an id from spot instance but EC2 doesn't recognize it yet; see above, eventual consistency possible
					continue
				}
				instance = instances.Reservations[0].Instances[0]
				err = nil
				break
			}
			time.Sleep(5 * time.Second)
		}

		if err != nil {
			return fmt.Errorf("Error resolving spot instance to real instance: %v", err)
		}
	} else {
		inst, err := d.getClient().RunInstances(&ec2.RunInstancesInput{
			ImageId:  &d.AMI,
			MinCount: aws.Int64(1),
			MaxCount: aws.Int64(1),
			Placement: &ec2.Placement{
				AvailabilityZone: &regionZone,
			},
			KeyName:           &d.KeyName,
			InstanceType:      &d.InstanceType,
			NetworkInterfaces: netSpecs,
			Monitoring:        &ec2.RunInstancesMonitoringEnabled{Enabled: aws.Bool(d.Monitoring)},
			IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
				Name: &d.IamInstanceProfile,
			},
			EbsOptimized:        &d.UseEbsOptimizedInstance,
			BlockDeviceMappings: []*ec2.BlockDeviceMapping{bdm},
		})

		if err != nil {
			return fmt.Errorf("Error launching instance: %s", err)
		}
		instance = inst.Instances[0]
	}

	d.InstanceId = *instance.InstanceId

	log.Debug("waiting for ip address to become available")
	if err := mcnutils.WaitFor(d.instanceIpAvailable); err != nil {
		return err
	}

	if instance.PrivateIpAddress != nil {
		d.PrivateIPAddress = *instance.PrivateIpAddress
	}

	d.waitForInstance()

	log.Debugf("created instance ID %s, IP address %s, Private IP address %s",
		d.InstanceId,
		d.IPAddress,
		d.PrivateIPAddress,
	)

	log.Debug("Settings tags for instance")
	err := d.configureTags(d.Tags)

	if err != nil {
		return fmt.Errorf("Unable to tag instance %s: %s", d.InstanceId, err)
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(dockerPort))), nil
}

func (d *Driver) GetIP() (string, error) {
	inst, err := d.getInstance()
	if err != nil {
		return "", err
	}

	if d.PrivateIPOnly {
		if inst.PrivateIpAddress == nil {
			return "", fmt.Errorf("No private IP for instance %v", *inst.InstanceId)
		}
		return *inst.PrivateIpAddress, nil
	}

	if d.UsePrivateIP {
		if inst.PrivateIpAddress == nil {
			return "", fmt.Errorf("No private IP for instance %v", *inst.InstanceId)
		}
		return *inst.PrivateIpAddress, nil
	}

	if inst.PublicIpAddress == nil {
		return "", fmt.Errorf("No IP for instance %v", *inst.InstanceId)
	}
	return *inst.PublicIpAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	inst, err := d.getInstance()
	if err != nil {
		return state.Error, err
	}
	switch *inst.State.Name {
	case ec2.InstanceStateNamePending:
		return state.Starting, nil
	case ec2.InstanceStateNameRunning:
		return state.Running, nil
	case ec2.InstanceStateNameStopping:
		return state.Stopping, nil
	case ec2.InstanceStateNameShuttingDown:
		return state.Stopping, nil
	case ec2.InstanceStateNameStopped:
		return state.Stopped, nil
	case ec2.InstanceStateNameTerminated:
		return state.Error, nil
	default:
		log.Warnf("unrecognized instance state: %v", *inst.State.Name)
		return state.Error, nil
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	// TODO: use @nathanleclaire retry func here (ehazlett)
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = defaultSSHUser
	}

	return d.SSHUser
}

func (d *Driver) Start() error {
	_, err := d.getClient().StartInstances(&ec2.StartInstancesInput{
		InstanceIds: []*string{&d.InstanceId},
	})
	if err != nil {
		return err
	}

	return d.waitForInstance()
}

func (d *Driver) Stop() error {
	_, err := d.getClient().StopInstances(&ec2.StopInstancesInput{
		InstanceIds: []*string{&d.InstanceId},
		Force:       aws.Bool(false),
	})
	return err
}

func (d *Driver) Restart() error {
	_, err := d.getClient().RebootInstances(&ec2.RebootInstancesInput{
		InstanceIds: []*string{&d.InstanceId},
	})
	return err
}

func (d *Driver) Kill() error {
	_, err := d.getClient().StopInstances(&ec2.StopInstancesInput{
		InstanceIds: []*string{&d.InstanceId},
		Force:       aws.Bool(true),
	})
	return err
}

func (d *Driver) Remove() error {
	if err := d.terminate(); err != nil {
		return fmt.Errorf("unable to terminate instance: %s", err)
	}

	// remove keypair
	if err := d.deleteKeyPair(); err != nil {
		return fmt.Errorf("unable to remove key pair: %s", err)
	}

	return nil
}

func (d *Driver) getInstance() (*ec2.Instance, error) {
	instances, err := d.getClient().DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{&d.InstanceId},
	})
	if err != nil {
		return nil, err
	}
	return instances.Reservations[0].Instances[0], nil
}

func (d *Driver) instanceIsRunning() bool {
	st, err := d.GetState()
	if err != nil {
		log.Debug(err)
	}
	if st == state.Running {
		return true
	}
	return false
}

func (d *Driver) waitForInstance() error {
	if err := mcnutils.WaitFor(d.instanceIsRunning); err != nil {
		return err
	}

	return nil
}

func (d *Driver) createKeyPair() error {

	keyPath := ""

	if d.SSHPrivateKeyPath == "" {
		log.Debugf("Creating New SSH Key")
		if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
			return err
		}
		keyPath = d.GetSSHKeyPath()
	} else {
		log.Debugf("Using ExistingKeyPair: %s", d.SSHPrivateKeyPath)
		if err := mcnutils.CopyFile(d.SSHPrivateKeyPath, d.GetSSHKeyPath()); err != nil {
			return err
		}
		if err := mcnutils.CopyFile(d.SSHPrivateKeyPath+".pub", d.GetSSHKeyPath()+".pub"); err != nil {
			return err
		}
		keyPath = d.SSHPrivateKeyPath
	}

	publicKey, err := ioutil.ReadFile(keyPath + ".pub")
	if err != nil {
		return err
	}

	keyName := d.MachineName

	log.Debugf("creating key pair: %s", keyName)
	_, err = d.getClient().ImportKeyPair(&ec2.ImportKeyPairInput{
		KeyName:           &keyName,
		PublicKeyMaterial: publicKey,
	})
	if err != nil {
		return err
	}
	d.KeyName = keyName
	return nil
}

func (d *Driver) terminate() error {
	if d.InstanceId == "" {
		return fmt.Errorf("unknown instance")
	}

	log.Debugf("terminating instance: %s", d.InstanceId)
	_, err := d.getClient().TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{&d.InstanceId},
	})
	if err != nil {
		return fmt.Errorf("unable to terminate instance: %s", err)
	}
	return nil
}

func (d *Driver) isSwarmMaster() bool {
	return d.SwarmMaster
}

func (d *Driver) securityGroupAvailableFunc(id string) func() bool {
	return func() bool {

		securityGroup, err := d.getClient().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{&id},
		})
		if err == nil && len(securityGroup.SecurityGroups) > 0 {
			return true
		} else if err == nil {
			log.Debugf("No security group with id %v found", id)
			return false
		}
		log.Debug(err)
		return false
	}
}

func (d *Driver) configureTags(tagGroups string) error {

	tags := []*ec2.Tag{}
	tags = append(tags, &ec2.Tag{
		Key:   aws.String("Name"),
		Value: &d.MachineName,
	})

	if tagGroups != "" {
		t := strings.Split(tagGroups, ",")
		if len(t) > 0 && len(t)%2 != 0 {
			log.Warnf("Tags are not key value in pairs. %d elements found", len(t))
		}
		for i := 0; i < len(t)-1; i += 2 {
			tags = append(tags, &ec2.Tag{
				Key:   &t[i],
				Value: &t[i+1],
			})
		}
	}

	_, err := d.getClient().CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{&d.InstanceId},
		Tags:      tags,
	})

	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) configureSecurityGroups(groupNames []string) error {
	if len(groupNames) == 0 {
		log.Debugf("no security groups to configure in %s", d.VpcId)
		return nil
	}

	log.Debugf("configuring security groups in %s", d.VpcId)

	filters := []*ec2.Filter{
		{
			Name:   aws.String("group-name"),
			Values: makePointerSlice(groupNames),
		},
		{
			Name:   aws.String("vpc-id"),
			Values: []*string{&d.VpcId},
		},
	}
	groups, err := d.getClient().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: filters,
	})
	if err != nil {
		return err
	}

	var groupsByName = make(map[string]*ec2.SecurityGroup)
	for _, securityGroup := range groups.SecurityGroups {
		groupsByName[*securityGroup.GroupName] = securityGroup
	}

	for _, groupName := range groupNames {
		var group *ec2.SecurityGroup
		securityGroup, ok := groupsByName[groupName]
		if ok {
			log.Debugf("found existing security group (%s) in %s", groupName, d.VpcId)
			group = securityGroup
		} else {
			log.Debugf("creating security group (%s) in %s", groupName, d.VpcId)
			groupResp, err := d.getClient().CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
				GroupName:   aws.String(groupName),
				Description: aws.String("Docker Machine"),
				VpcId:       aws.String(d.VpcId),
			})
			if err != nil {
				return err
			}
			// Manually translate into the security group construct
			group = &ec2.SecurityGroup{
				GroupId:   groupResp.GroupId,
				VpcId:     aws.String(d.VpcId),
				GroupName: aws.String(groupName),
			}
			// wait until created (dat eventual consistency)
			log.Debugf("waiting for group (%s) to become available", *group.GroupId)
			if err := mcnutils.WaitFor(d.securityGroupAvailableFunc(*group.GroupId)); err != nil {
				return err
			}
		}
		d.SecurityGroupIds = append(d.SecurityGroupIds, *group.GroupId)

		perms := d.configureSecurityGroupPermissions(group)

		if len(perms) != 0 {
			log.Debugf("authorizing group %s with permissions: %v", groupNames, perms)
			_, err := d.getClient().AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
				GroupId:       group.GroupId,
				IpPermissions: perms,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Driver) configureSecurityGroupPermissions(group *ec2.SecurityGroup) []*ec2.IpPermission {
	hasSshPort := false
	hasDockerPort := false
	hasSwarmPort := false
	for _, p := range group.IpPermissions {
		if p.FromPort != nil {
			switch *p.FromPort {
			case 22:
				hasSshPort = true
			case int64(dockerPort):
				hasDockerPort = true
			case int64(swarmPort):
				hasSwarmPort = true
			}
		}
	}

	perms := []*ec2.IpPermission{}

	if !hasSshPort {
		perms = append(perms, &ec2.IpPermission{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(22),
			ToPort:     aws.Int64(22),
			IpRanges:   []*ec2.IpRange{{CidrIp: aws.String(ipRange)}},
		})
	}

	if !hasDockerPort {
		perms = append(perms, &ec2.IpPermission{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(int64(dockerPort)),
			ToPort:     aws.Int64(int64(dockerPort)),
			IpRanges:   []*ec2.IpRange{{CidrIp: aws.String(ipRange)}},
		})
	}

	if !hasSwarmPort && d.SwarmMaster {
		perms = append(perms, &ec2.IpPermission{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(int64(swarmPort)),
			ToPort:     aws.Int64(int64(swarmPort)),
			IpRanges:   []*ec2.IpRange{{CidrIp: aws.String(ipRange)}},
		})
	}

	log.Debugf("configuring security group authorization for %s", ipRange)

	return perms
}

func (d *Driver) deleteKeyPair() error {
	log.Debugf("deleting key pair: %s", d.KeyName)

	_, err := d.getClient().DeleteKeyPair(&ec2.DeleteKeyPairInput{
		KeyName: &d.KeyName,
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) getDefaultVPCId() (string, error) {
	output, err := d.getClient().DescribeAccountAttributes(&ec2.DescribeAccountAttributesInput{})
	if err != nil {
		return "", err
	}

	for _, attribute := range output.AccountAttributes {
		if *attribute.AttributeName == "default-vpc" {
			return *attribute.AttributeValues[0].AttributeValue, nil
		}
	}

	return "", errors.New("No default-vpc attribute")
}

func generateId() string {
	rb := make([]byte, 10)
	_, err := rand.Read(rb)
	if err != nil {
		log.Warnf("Unable to generate id: %s", err)
	}

	h := md5.New()
	io.WriteString(h, string(rb))
	return fmt.Sprintf("%x", h.Sum(nil))
}
