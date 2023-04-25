package azureutil

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/minikube/pkg/libmachine/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/libmachine/mcnutils"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/minikube/pkg/libmachine/drivers/azure/logutil"
)

// Azure driver allows two authentication methods:
//
// 1. OAuth Device Flow
//
// Azure Active Directory implements OAuth 2.0 Device Flow described here:
// https://tools.ietf.org/html/draft-denniss-oauth-device-flow-00. It is simple
// for users to authenticate through a browser and requires re-authenticating
// every 2 weeks.
//
// Device auth prints a message to the screen telling the user to click on URL
// and approve the app on the browser, meanwhile the client polls the auth API
// for a token. Once we have token, we save it locally to a file with proper
// permissions and when the token expires (in Azure case typically 1 hour) SDK
// will automatically refresh the specified token and will call the refresh
// callback function we implement here. This way we will always be storing a
// token with a refresh_token saved on the machine.
//
// 2. Azure Service Principal Account
//
// This is designed for headless authentication to Azure APIs but requires more
// steps from user to create a Service Principal Account and provide its
// credentials to the machine driver.

var (
	// AD app id for docker-machine driver in various Azure realms
	appIDs = map[string]string{
		azure.PublicCloud.Name: "637ddaba-219b-43b8-bf19-8cea500cf273",
		azure.ChinaCloud.Name:  "bb5eed6f-120b-4365-8fd9-ab1a3fba5698",
		azure.GermanCloud.Name: "aabac5f7-dd47-47ef-824c-e0d57598cada",
	}
)

// AuthenticateDeviceFlow fetches a token from the local file cache or initiates a consent
// flow and waits for token to be obtained. Obtained token is stored in a file cache for
// future use and refreshing.
func AuthenticateDeviceFlow(env azure.Environment, subscriptionID string) (*azure.ServicePrincipalToken, error) {
	// First we locate the tenant ID of the subscription as we store tokens per
	// tenant (which could have multiple subscriptions)
	tenantID, err := loadOrFindTenantID(env, subscriptionID)
	if err != nil {
		return nil, err
	}
	oauthCfg, err := env.OAuthConfigForTenant(tenantID)
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain oauth config for azure environment: %v", err)
	}

	tokenPath := tokenCachePath(tenantID)
	saveToken := mkTokenCallback(tokenPath)
	saveTokenCallback := func(t azure.Token) error {
		log.Debug("Azure token expired. Saving the refreshed token...")
		return saveToken(t)
	}
	f := logutil.Fields{"path": tokenPath}

	appID, ok := appIDs[env.Name]
	if !ok {
		return nil, fmt.Errorf("docker-machine application not set up for Azure environment %q", env.Name)
	}
	scope := getScope(env)

	// Lookup the token cache file for an existing token.
	spt, err := tokenFromFile(*oauthCfg, tokenPath, appID, scope, saveTokenCallback)
	if err != nil {
		return nil, err
	}
	if spt != nil {
		log.Debug("Auth token found in file.", f)

		// NOTE(ahmetalpbalkan): The token file we found might be containing an
		// expired access_token. In that case, the first call to Azure SDK will
		// attempt to refresh the token using refresh_token –which might have
		// expired[1], in that case we will get an error and we shall remove the
		// token file and initiate token flow again so that the user would not
		// need removing the token cache file manually.
		//
		// [1]: for device flow auth, the expiration date of refresh_token is
		//      not returned in AAD /token response, we just know it is 14
		//      days. Therefore user’s token will go stale every 14 days and we
		//      will delete the token file, re-initiate the device flow. Service
		//      Principal Account tokens are not subject to this limitation.
		log.Debug("Validating the token.")
		if err := validateToken(env, spt); err != nil {
			log.Debug(fmt.Sprintf("Error: %v", err))
			log.Debug(fmt.Sprintf("Deleting %s", tokenPath))
			if err := os.RemoveAll(tokenPath); err != nil {
				return nil, fmt.Errorf("Error deleting stale token file: %v", err)
			}
		} else {
			log.Debug("Token works.")
			return spt, nil
		}
	}

	log.Debug("Obtaining a token.", f)
	spt, err = deviceFlowAuth(*oauthCfg, appID, scope)
	if err != nil {
		return nil, err
	}
	log.Debug("Obtained a token.")
	if err := saveToken(spt.Token); err != nil {
		log.Error("Error occurred saving token to cache file.")
		return nil, err
	}
	return spt, nil
}

// AuthenticateServicePrincipal uses given service principal credentials to return a
// service principal token. Generated token is not stored in a cache file or refreshed.
func AuthenticateServicePrincipal(env azure.Environment, subscriptionID, spID, spPassword string) (*azure.ServicePrincipalToken, error) {
	tenantID, err := loadOrFindTenantID(env, subscriptionID)
	if err != nil {
		return nil, err
	}
	oauthCfg, err := env.OAuthConfigForTenant(tenantID)
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain oauth config for azure environment: %v", err)
	}

	spt, err := azure.NewServicePrincipalToken(*oauthCfg, spID, spPassword, getScope(env))
	if err != nil {
		return nil, fmt.Errorf("Failed to create service principal token: %+v", err)
	}
	return spt, nil
}

// tokenFromFile returns a token from the specified file if it is found, otherwise
// returns nil. Any error retrieving or creating the token is returned as an error.
func tokenFromFile(oauthCfg azure.OAuthConfig, tokenPath, clientID, resource string,
	callback azure.TokenRefreshCallback) (*azure.ServicePrincipalToken, error) {
	log.Debug("Loading auth token from file", logutil.Fields{"path": tokenPath})
	if _, err := os.Stat(tokenPath); err != nil {
		if os.IsNotExist(err) { // file not found
			return nil, nil
		}
		return nil, err
	}

	token, err := azure.LoadToken(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load token from file: %v", err)
	}

	spt, err := azure.NewServicePrincipalTokenFromManualToken(oauthCfg, clientID, resource, *token, callback)
	if err != nil {
		return nil, fmt.Errorf("Error constructing service principal token: %v", err)
	}
	return spt, nil
}

// deviceFlowAuth prints a message to the screen for user to take action to
// consent application on a browser and in the meanwhile the authentication
// endpoint is polled until user gives consent, denies or the flow times out.
// Returned token must be saved.
func deviceFlowAuth(oauthCfg azure.OAuthConfig, clientID, resource string) (*azure.ServicePrincipalToken, error) {
	cl := oauthClient()
	deviceCode, err := azure.InitiateDeviceAuth(&cl, oauthCfg, clientID, resource)
	if err != nil {
		return nil, fmt.Errorf("Failed to start device auth: %v", err)
	}
	log.Debug("Retrieved device code.", logutil.Fields{
		"expires_in": to.Int64(deviceCode.ExpiresIn),
		"interval":   to.Int64(deviceCode.Interval),
	})

	// Example message: “To sign in, open https://aka.ms/devicelogin and enter
	// the code 0000000 to authenticate.”
	log.Infof("Microsoft Azure: %s", to.String(deviceCode.Message))

	token, err := azure.WaitForUserCompletion(&cl, deviceCode)
	if err != nil {
		return nil, fmt.Errorf("Failed to complete device auth: %v", err)
	}

	spt, err := azure.NewServicePrincipalTokenFromManualToken(oauthCfg, clientID, resource, *token)
	if err != nil {
		return nil, fmt.Errorf("Error constructing service principal token: %v", err)
	}
	return spt, nil
}

// azureCredsPath returns the directory the azure credentials are stored in.
func azureCredsPath() string {
	return filepath.Join(mcnutils.GetHomeDir(), ".docker", "machine", "credentials", "azure")
}

// tokenCachePath returns the full path the OAuth 2.0 token should be saved at
// for given tenant ID.
func tokenCachePath(tenantID string) string {
	return filepath.Join(azureCredsPath(), fmt.Sprintf("%s.json", tenantID))
}

// tenantIDPath returns the full path the tenant ID for the given subscription
// should be saved at.f
func tenantIDPath(subscriptionID string) string {
	return filepath.Join(azureCredsPath(), fmt.Sprintf("%s.tenantid", subscriptionID))
}

// mkTokenCallback returns a callback function that can be used to save the
// token initially or register to the Azure SDK to be called when the token is
// refreshed.
func mkTokenCallback(path string) azure.TokenRefreshCallback {
	return func(t azure.Token) error {
		if err := azure.SaveToken(path, 0600, t); err != nil {
			return err
		}
		log.Debug("Saved token to file.")
		return nil
	}
}

// validateToken makes a call to Azure SDK with given token, essentially making
// sure if the access_token valid, if not it uses SDK’s functionality to
// automatically refresh the token using refresh_token (which might have
// expired). This check is essentially to make sure refresh_token is good.
func validateToken(env azure.Environment, token *azure.ServicePrincipalToken) error {
	c := subscriptionsClient(env.ResourceManagerEndpoint)
	c.Authorizer = token
	_, err := c.List()
	if err != nil {
		return fmt.Errorf("Token validity check failed: %v", err)
	}
	return nil
}

// getScope returns the API scope for authentication tokens.
func getScope(env azure.Environment) string {
	// for AzurePublicCloud (https://management.core.windows.net/), this old
	// Service Management scope covers both ASM and ARM.
	return env.ServiceManagementEndpoint
}
