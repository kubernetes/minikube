package amazonec2

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type awsCredentials interface {
	Credentials() *credentials.Credentials
}

type ProviderFactory interface {
	NewStaticProvider(id, secret, token string) credentials.Provider
}

type defaultAWSCredentials struct {
	AccessKey        string
	SecretKey        string
	SessionToken     string
	providerFactory  ProviderFactory
	fallbackProvider awsCredentials
}

func NewAWSCredentials(id, secret, token string) *defaultAWSCredentials {
	creds := defaultAWSCredentials{
		AccessKey:        id,
		SecretKey:        secret,
		SessionToken:     token,
		fallbackProvider: &AwsDefaultCredentialsProvider{},
		providerFactory:  &defaultProviderFactory{},
	}
	return &creds
}

func (c *defaultAWSCredentials) Credentials() *credentials.Credentials {
	providers := []credentials.Provider{}
	if c.AccessKey != "" && c.SecretKey != "" {
		providers = append(providers, c.providerFactory.NewStaticProvider(c.AccessKey, c.SecretKey, c.SessionToken))
	}
	if c.fallbackProvider != nil {
		fallbackCreds, err := c.fallbackProvider.Credentials().Get()
		if err == nil {
			providers = append(providers, &credentials.StaticProvider{Value: fallbackCreds})
		}
	}
	return credentials.NewChainCredentials(providers)
}

type AwsDefaultCredentialsProvider struct{}

func (c *AwsDefaultCredentialsProvider) Credentials() *credentials.Credentials {
	return session.New().Config.Credentials
}

type defaultProviderFactory struct{}

func (c *defaultProviderFactory) NewStaticProvider(id, secret, token string) credentials.Provider {
	return &credentials.StaticProvider{Value: credentials.Value{
		AccessKeyID:     id,
		SecretAccessKey: secret,
		SessionToken:    token,
	}}
}
