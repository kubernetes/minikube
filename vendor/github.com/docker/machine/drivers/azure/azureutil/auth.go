package azureutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/docker/machine/drivers/azure/logutil"
)

var (
	// AD app id for docker-machine driver in various Azure realms
	clientIDs = map[string]string{
		azure.PublicCloud.Name: "637ddaba-219b-43b8-bf19-8cea500cf273",
		azure.ChinaCloud.Name:  "bb5eed6f-120b-4365-8fd9-ab1a3fba5698",
	}
)

// NOTE(ahmetalpbalkan): Azure Active Directory implements OAuth 2.0 Device Flow
// described here: https://tools.ietf.org/html/draft-denniss-oauth-device-flow-00
// Although it has some gotchas, most of the authentication logic is in Azure SDK
// for Go helper packages.
//
// Device auth prints a message to the screen telling the user to click on URL
// and approve the app on the browser, meanwhile the client polls the auth API
// for a token. Once we have token, we save it locally to a file with proper
// permissions and when the token expires (in Azure case typically 1 hour) SDK
// will automatically refresh the specified token and will call the refresh
// callback function we implement here. This way we will always be storing a
// token with a refresh_token saved on the machine.

// Authenticate fetches a token from the local file cache or initiates a consent
// flow and waits for token to be obtained.
func Authenticate(env azure.Environment, subscriptionID string) (*azure.ServicePrincipalToken, error) {
	clientID, ok := clientIDs[env.Name]
	if !ok {
		return nil, fmt.Errorf("docker-machine application not set up for Azure environment %q", env.Name)
	}

	// First we locate the tenant ID of the subscription as we store tokens per
	// tenant (which could have multiple subscriptions)
	log.Debug("Looking up AAD Tenant ID.", logutil.Fields{
		"subs": subscriptionID})
	tenantID, err := loadOrFindTenantID(env, subscriptionID)
	if err != nil {
		return nil, err
	}
	log.Debug("Found AAD Tenant ID.", logutil.Fields{
		"tenant": tenantID,
		"subs":   subscriptionID})

	oauthCfg, err := env.OAuthConfigForTenant(tenantID)
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain oauth config for azure environment: %v", err)
	}

	// for AzurePublicCloud (https://management.core.windows.net/), this old
	// Service Management scope covers both ASM and ARM.
	apiScope := env.ServiceManagementEndpoint

	tokenPath := tokenCachePath(tenantID)
	saveToken := mkTokenCallback(tokenPath)
	saveTokenCallback := func(t azure.Token) error {
		log.Debug("Azure token expired. Saving the refreshed token...")
		return saveToken(t)
	}
	f := logutil.Fields{"path": tokenPath}

	// Lookup the token cache file for an existing token.
	spt, err := tokenFromFile(*oauthCfg, tokenPath, clientID, apiScope, saveTokenCallback)
	if err != nil {
		return nil, err
	}
	if spt != nil {
		log.Debug("Auth token found in file.", f)

		// NOTE(ahmetalpbalkan): The token file we found might be containng an
		// expired access_token. In that case, the first call to Azure SDK will
		// attempt to refresh the token using refresh_token –which might have
		// expired[1], in that case we will get an error and we shall remove the
		// token file and initiate token flow again so that the user would not
		// need removing the token cache file manually.
		//
		// [1]: expiration date of refresh_token is not returned in AAD /token
		//      response, we just know it is 14 days. Therefore user’s token
		//      will go stale every 14 days and we will delete the token file,
		//      re-initiate the device flow.
		log.Debug("Validating the token.")
		if err := validateToken(env, spt); err != nil {
			log.Debug(fmt.Sprintf("Error: %v", err))
			log.Info("Stored Azure credentials expired. Please reauthenticate.")
			log.Debug(fmt.Sprintf("Deleting %s", tokenPath))
			if err := os.RemoveAll(tokenPath); err != nil {
				return nil, fmt.Errorf("Error deleting stale token file: %v", err)
			}
		} else {
			log.Debug("Token works.")
			return spt, nil
		}
	}

	// Start an OAuth 2.0 device flow
	log.Debug("Initiating device flow.", f)
	spt, err = tokenFromDeviceFlow(*oauthCfg, tokenPath, clientID, apiScope)
	if err != nil {
		return nil, err
	}
	log.Debug("Obtained service principal token.")
	if err := saveToken(spt.Token); err != nil {
		log.Error("Error occurred saving token to cache file.")
		return nil, err
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

// tokenFromDeviceFlow prints a message to the screen for user to take action to
// consent application on a browser and in the meanwhile the authentication
// endpoint is polled until user gives consent, denies or the flow times out.
// Returned token must be saved.
func tokenFromDeviceFlow(oauthCfg azure.OAuthConfig, tokenPath, clientID, resource string) (*azure.ServicePrincipalToken, error) {
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
// should be saved at.
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
