// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsconfig

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

// Copied from https://github.com/aws/aws-sdk-go-v2/blob/main/internal/configsources/config.go
type UseFIPSEndpointProvider interface {
	GetUseFIPSEndpoint(context.Context) (value aws.FIPSEndpointState, found bool, err error)
}

// Copied from https://github.com/aws/aws-sdk-go-v2/blob/main/internal/configsources/config.go
func ResolveUseFIPSEndpoint(ctx context.Context, configSources []any) (value aws.FIPSEndpointState, found bool, err error) {
	for _, cfg := range configSources {
		if p, ok := cfg.(UseFIPSEndpointProvider); ok {
			value, found, err = p.GetUseFIPSEndpoint(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

func FIPSEndpointStateString(state aws.FIPSEndpointState) string {
	switch state {
	case aws.FIPSEndpointStateUnset:
		return "FIPSEndpointStateUnset"
	case aws.FIPSEndpointStateEnabled:
		return "FIPSEndpointStateEnabled"
	case aws.FIPSEndpointStateDisabled:
		return "FIPSEndpointStateDisabled"
	}
	return fmt.Sprintf("unknown aws.FIPSEndpointState (%d)", state)
}

// Copied from https://github.com/aws/aws-sdk-go-v2/blob/main/internal/configsources/config.go
type UseDualStackEndpointProvider interface {
	GetUseDualStackEndpoint(context.Context) (value aws.DualStackEndpointState, found bool, err error)
}

// Copied from https://github.com/aws/aws-sdk-go-v2/blob/main/internal/configsources/config.go
func ResolveUseDualStackEndpoint(ctx context.Context, configSources []any) (value aws.DualStackEndpointState, found bool, err error) {
	for _, cfg := range configSources {
		if p, ok := cfg.(UseDualStackEndpointProvider); ok {
			value, found, err = p.GetUseDualStackEndpoint(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return
}

func DualStackEndpointStateString(state aws.DualStackEndpointState) string {
	switch state {
	case aws.DualStackEndpointStateUnset:
		return "DualStackEndpointStateUnset"
	case aws.DualStackEndpointStateEnabled:
		return "DualStackEndpointStateEnabled"
	case aws.DualStackEndpointStateDisabled:
		return "DualStackEndpointStateDisabled"
	}
	return fmt.Sprintf("unknown aws.FIPSEndpointStateUnset (%d)", state)
}

// Copied and renamed from https://github.com/aws/aws-sdk-go-v2/blob/main/feature/ec2/imds/internal/config/resolvers.go
type EC2IMDSClientEnableStateResolver interface {
	GetEC2IMDSClientEnableState() (imds.ClientEnableState, bool, error)
}

// Copied and renamed from https://github.com/aws/aws-sdk-go-v2/blob/main/feature/ec2/imds/internal/config/resolvers.go
func ResolveEC2IMDSClientEnableState(sources []any) (value imds.ClientEnableState, found bool, err error) {
	for _, source := range sources {
		if resolver, ok := source.(EC2IMDSClientEnableStateResolver); ok {
			value, found, err = resolver.GetEC2IMDSClientEnableState()
			if err != nil || found {
				return value, found, err
			}
		}
	}
	return value, found, err
}

func EC2IMDSClientEnableStateString(state imds.ClientEnableState) string {
	switch state {
	case imds.ClientDefaultEnableState:
		return "ClientDefaultEnableState"
	case imds.ClientDisabled:
		return "ClientDisabled"
	case imds.ClientEnabled:
		return "ClientEnabled"
	}
	return fmt.Sprintf("unknown imds.ClientEnableState (%d)", state)
}

// Copied and renamed from https://github.com/aws/aws-sdk-go-v2/blob/main/feature/ec2/imds/internal/config/resolvers.go
type EC2IMDSEndpointResolver interface {
	GetEC2IMDSEndpoint() (value string, found bool, err error)
}

// Copied and renamed from https://github.com/aws/aws-sdk-go-v2/blob/main/feature/ec2/imds/internal/config/resolvers.go
func ResolveEC2IMDSEndpointConfig(configSources []any) (value string, found bool, err error) {
	for _, cfg := range configSources {
		if p, ok := cfg.(EC2IMDSEndpointResolver); ok {
			value, found, err = p.GetEC2IMDSEndpoint()
			if err != nil || found {
				break
			}
		}
	}
	return
}

// Copied and renamed from https://github.com/aws/aws-sdk-go-v2/blob/main/feature/ec2/imds/internal/config/resolvers.go
type EC2IMDSEndpointModeResolver interface {
	GetEC2IMDSEndpointMode() (imds.EndpointModeState, bool, error)
}

// Copied and renamed from https://github.com/aws/aws-sdk-go-v2/blob/main/feature/ec2/imds/internal/config/resolvers.go
func ResolveEC2IMDSEndpointModeConfig(sources []any) (value imds.EndpointModeState, found bool, err error) {
	for _, source := range sources {
		if resolver, ok := source.(EC2IMDSEndpointModeResolver); ok {
			value, found, err = resolver.GetEC2IMDSEndpointMode()
			if err != nil || found {
				return value, found, err
			}
		}
	}
	return value, found, err
}

func EC2IMDSEndpointModeString(state imds.EndpointModeState) string {
	switch state {
	case imds.EndpointModeStateUnset:
		return "EndpointModeStateUnset"
	case imds.EndpointModeStateIPv4:
		return "EndpointModeStateIPv4"
	case imds.EndpointModeStateIPv6:
		return "EndpointModeStateIPv6"
	}
	return fmt.Sprintf("unknown imds.EndpointModeState (%d)", state)
}

// Copied and renamed from https://github.com/aws/aws-sdk-go-v2/blob/main/config/provider.go
type RetryMaxAttemptsProvider interface {
	GetRetryMaxAttempts(context.Context) (int, bool, error)
}

// Copied and renamed from https://github.com/aws/aws-sdk-go-v2/blob/main/config/provider.go
func GetRetryMaxAttempts(ctx context.Context, sources []any) (v int, found bool, err error) {
	for _, c := range sources {
		if p, ok := c.(RetryMaxAttemptsProvider); ok {
			v, found, err = p.GetRetryMaxAttempts(ctx)
			if err != nil || found {
				break
			}
		}
	}
	return v, found, err
}
