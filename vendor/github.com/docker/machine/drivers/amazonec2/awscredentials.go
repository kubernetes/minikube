package amazonec2

import "github.com/aws/aws-sdk-go/aws/credentials"

type awsCredentials interface {
	NewStaticCredentials(id, secret, token string) *credentials.Credentials

	NewSharedCredentials(filename, profile string) *credentials.Credentials
}

type defaultAWSCredentials struct{}

func (c *defaultAWSCredentials) NewStaticCredentials(id, secret, token string) *credentials.Credentials {
	return credentials.NewStaticCredentials(id, secret, token)
}

func (c *defaultAWSCredentials) NewSharedCredentials(filename, profile string) *credentials.Credentials {
	return credentials.NewSharedCredentials(filename, profile)
}
