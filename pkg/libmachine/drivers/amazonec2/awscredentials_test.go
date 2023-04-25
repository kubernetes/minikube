package amazonec2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccessKeyIsMandatoryWhenSystemCredentialsAreNotPresent(t *testing.T) {
	awsCreds := NewAWSCredentials("", "", "")
	awsCreds.fallbackProvider = nil

	_, err := awsCreds.Credentials().Get()
	assert.Error(t, err)
}

func TestAccessKeyIsMandatoryEvenIfSecretKeyIsPassedWhenSystemCredentialsAreNotPresent(t *testing.T) {
	awsCreds := NewAWSCredentials("", "secret", "")
	awsCreds.fallbackProvider = nil

	_, err := awsCreds.Credentials().Get()
	assert.Error(t, err)
}

func TestSecretKeyIsMandatoryWhenSystemCredentialsAreNotPresent(t *testing.T) {
	awsCreds := NewAWSCredentials("access", "", "")
	awsCreds.fallbackProvider = nil

	_, err := awsCreds.Credentials().Get()
	assert.Error(t, err)
}

func TestFallbackCredentialsAreLoadedWhenAccessKeyAndSecretKeyAreMissing(t *testing.T) {
	awsCreds := NewAWSCredentials("", "", "")
	awsCreds.fallbackProvider = &fallbackCredentials{}

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "fallback_access", creds.AccessKeyID)
	assert.Equal(t, "fallback_secret", creds.SecretAccessKey)
	assert.Equal(t, "fallback_token", creds.SessionToken)
}

func TestFallbackCredentialsAreLoadedWhenAccessKeyIsMissing(t *testing.T) {
	awsCreds := NewAWSCredentials("", "secret", "")
	awsCreds.fallbackProvider = &fallbackCredentials{}

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "fallback_access", creds.AccessKeyID)
	assert.Equal(t, "fallback_secret", creds.SecretAccessKey)
	assert.Equal(t, "fallback_token", creds.SessionToken)
}

func TestFallbackCredentialsAreLoadedWhenSecretKeyIsMissing(t *testing.T) {
	awsCreds := NewAWSCredentials("access", "", "")
	awsCreds.fallbackProvider = &fallbackCredentials{}

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "fallback_access", creds.AccessKeyID)
	assert.Equal(t, "fallback_secret", creds.SecretAccessKey)
	assert.Equal(t, "fallback_token", creds.SessionToken)
}

func TestOptionCredentialsAreLoadedWhenAccessKeyAndSecretKeyAreProvided(t *testing.T) {
	awsCreds := NewAWSCredentials("access", "secret", "")
	awsCreds.fallbackProvider = &fallbackCredentials{}

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "access", creds.AccessKeyID)
	assert.Equal(t, "secret", creds.SecretAccessKey)
	assert.Equal(t, "", creds.SessionToken)
}

func TestFallbackCredentialsAreLoadedIfStaticCredentialsGenerateError(t *testing.T) {
	awsCreds := NewAWSCredentials("access", "secret", "token")
	awsCreds.fallbackProvider = &fallbackCredentials{}
	awsCreds.providerFactory = &errorCredentialsProvider{}

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "fallback_access", creds.AccessKeyID)
	assert.Equal(t, "fallback_secret", creds.SecretAccessKey)
	assert.Equal(t, "fallback_token", creds.SessionToken)
}

func TestErrorGeneratedWhenAllProvidersGenerateErrors(t *testing.T) {
	awsCreds := NewAWSCredentials("access", "secret", "token")
	awsCreds.fallbackProvider = &errorFallbackCredentials{}
	awsCreds.providerFactory = &errorCredentialsProvider{}

	_, err := awsCreds.Credentials().Get()
	assert.Error(t, err)
}
