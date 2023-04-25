package azureutil

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest"
)

// accessToken is interim autorest.Authorizer until we figure out oauth token
// handling. It holds the access token.
type accessToken string

func (a accessToken) WithAuthorization() autorest.PrepareDecorator {
	return autorest.WithHeader("Authorization", fmt.Sprintf("Bearer %s", string(a)))
}
