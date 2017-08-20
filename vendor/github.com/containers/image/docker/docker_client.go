package docker

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/containers/image/docker/reference"
	"github.com/containers/image/types"
	"github.com/containers/storage/pkg/homedir"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

const (
	dockerHostname     = "docker.io"
	dockerRegistry     = "registry-1.docker.io"
	dockerAuthRegistry = "https://index.docker.io/v1/"

	dockerCfg         = ".docker"
	dockerCfgFileName = "config.json"
	dockerCfgObsolete = ".dockercfg"

	systemPerHostCertDirPath = "/etc/docker/certs.d"

	resolvedPingV2URL       = "%s://%s/v2/"
	resolvedPingV1URL       = "%s://%s/v1/_ping"
	tagsPath                = "/v2/%s/tags/list"
	manifestPath            = "/v2/%s/manifests/%s"
	blobsPath               = "/v2/%s/blobs/%s"
	blobUploadPath          = "/v2/%s/blobs/uploads/"
	extensionsSignaturePath = "/extensions/v2/%s/signatures/%s"

	minimumTokenLifetimeSeconds = 60

	extensionSignatureSchemaVersion = 2        // extensionSignature.Version
	extensionSignatureTypeAtomic    = "atomic" // extensionSignature.Type
)

// ErrV1NotSupported is returned when we're trying to talk to a
// docker V1 registry.
var ErrV1NotSupported = errors.New("can't talk to a V1 docker registry")

// extensionSignature and extensionSignatureList come from github.com/openshift/origin/pkg/dockerregistry/server/signaturedispatcher.go:
// signature represents a Docker image signature.
type extensionSignature struct {
	Version int    `json:"schemaVersion"` // Version specifies the schema version
	Name    string `json:"name"`          // Name must be in "sha256:<digest>@signatureName" format
	Type    string `json:"type"`          // Type is optional, of not set it will be defaulted to "AtomicImageV1"
	Content []byte `json:"content"`       // Content contains the signature
}

// signatureList represents list of Docker image signatures.
type extensionSignatureList struct {
	Signatures []extensionSignature `json:"signatures"`
}

type bearerToken struct {
	Token     string    `json:"token"`
	ExpiresIn int       `json:"expires_in"`
	IssuedAt  time.Time `json:"issued_at"`
}

// dockerClient is configuration for dealing with a single Docker registry.
type dockerClient struct {
	// The following members are set by newDockerClient and do not change afterwards.
	ctx           *types.SystemContext
	registry      string
	username      string
	password      string
	client        *http.Client
	signatureBase signatureStorageBase
	scope         authScope
	// The following members are detected registry properties:
	// They are set after a successful detectProperties(), and never change afterwards.
	scheme             string // Empty value also used to indicate detectProperties() has not yet succeeded.
	challenges         []challenge
	supportsSignatures bool
	// The following members are private state for setupRequestAuth, both are valid if token != nil.
	token           *bearerToken
	tokenExpiration time.Time
}

type authScope struct {
	remoteName string
	actions    string
}

// this is cloned from docker/go-connections because upstream docker has changed
// it and make deps here fails otherwise.
// We'll drop this once we upgrade to docker 1.13.x deps.
func serverDefault() *tls.Config {
	return &tls.Config{
		// Avoid fallback to SSL protocols < TLS1.0
		MinVersion:               tls.VersionTLS10,
		PreferServerCipherSuites: true,
		CipherSuites:             tlsconfig.DefaultServerAcceptedCiphers,
	}
}

func newTransport() *http.Transport {
	direct := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	tr := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                direct.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		// TODO(dmcgowan): Call close idle connections when complete and use keep alive
		DisableKeepAlives: true,
	}
	proxyDialer, err := sockets.DialerFromEnvironment(direct)
	if err == nil {
		tr.Dial = proxyDialer.Dial
	}
	return tr
}

// dockerCertDir returns a path to a directory to be consumed by setupCertificates() depending on ctx and hostPort.
func dockerCertDir(ctx *types.SystemContext, hostPort string) string {
	if ctx != nil && ctx.DockerCertPath != "" {
		return ctx.DockerCertPath
	}
	var hostCertDir string
	if ctx != nil && ctx.DockerPerHostCertDirPath != "" {
		hostCertDir = ctx.DockerPerHostCertDirPath
	} else if ctx != nil && ctx.RootForImplicitAbsolutePaths != "" {
		hostCertDir = filepath.Join(ctx.RootForImplicitAbsolutePaths, systemPerHostCertDirPath)
	} else {
		hostCertDir = systemPerHostCertDirPath
	}
	return filepath.Join(hostCertDir, hostPort)
}

func setupCertificates(dir string, tlsc *tls.Config) error {
	logrus.Debugf("Looking for TLS certificates and private keys in %s", dir)
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, f := range fs {
		fullPath := filepath.Join(dir, f.Name())
		if strings.HasSuffix(f.Name(), ".crt") {
			systemPool, err := tlsconfig.SystemCertPool()
			if err != nil {
				return errors.Wrap(err, "unable to get system cert pool")
			}
			tlsc.RootCAs = systemPool
			logrus.Debugf(" crt: %s", fullPath)
			data, err := ioutil.ReadFile(fullPath)
			if err != nil {
				return err
			}
			tlsc.RootCAs.AppendCertsFromPEM(data)
		}
		if strings.HasSuffix(f.Name(), ".cert") {
			certName := f.Name()
			keyName := certName[:len(certName)-5] + ".key"
			logrus.Debugf(" cert: %s", fullPath)
			if !hasFile(fs, keyName) {
				return errors.Errorf("missing key %s for client certificate %s. Note that CA certificates should use the extension .crt", keyName, certName)
			}
			cert, err := tls.LoadX509KeyPair(filepath.Join(dir, certName), filepath.Join(dir, keyName))
			if err != nil {
				return err
			}
			tlsc.Certificates = append(tlsc.Certificates, cert)
		}
		if strings.HasSuffix(f.Name(), ".key") {
			keyName := f.Name()
			certName := keyName[:len(keyName)-4] + ".cert"
			logrus.Debugf(" key: %s", fullPath)
			if !hasFile(fs, certName) {
				return errors.Errorf("missing client certificate %s for key %s", certName, keyName)
			}
		}
	}
	return nil
}

func hasFile(files []os.FileInfo, name string) bool {
	for _, f := range files {
		if f.Name() == name {
			return true
		}
	}
	return false
}

// newDockerClient returns a new dockerClient instance for refHostname (a host a specified in the Docker image reference, not canonicalized to dockerRegistry)
// “write” specifies whether the client will be used for "write" access (in particular passed to lookaside.go:toplevelFromSection)
func newDockerClient(ctx *types.SystemContext, ref dockerReference, write bool, actions string) (*dockerClient, error) {
	registry := reference.Domain(ref.ref)
	if registry == dockerHostname {
		registry = dockerRegistry
	}
	username, password, err := getAuth(ctx, reference.Domain(ref.ref))
	if err != nil {
		return nil, err
	}
	tr := newTransport()
	tr.TLSClientConfig = serverDefault()
	// It is undefined whether the host[:port] string for dockerHostname should be dockerHostname or dockerRegistry,
	// because docker/docker does not read the certs.d subdirectory at all in that case.  We use the user-visible
	// dockerHostname here, because it is more symmetrical to read the configuration in that case as well, and because
	// generally the UI hides the existence of the different dockerRegistry.  But note that this behavior is
	// undocumented and may change if docker/docker changes.
	certDir := dockerCertDir(ctx, reference.Domain(ref.ref))
	if err := setupCertificates(certDir, tr.TLSClientConfig); err != nil {
		return nil, err
	}
	if ctx != nil && ctx.DockerInsecureSkipTLSVerify {
		tr.TLSClientConfig.InsecureSkipVerify = true
	}
	client := &http.Client{Transport: tr}

	sigBase, err := configuredSignatureStorageBase(ctx, ref, write)
	if err != nil {
		return nil, err
	}

	return &dockerClient{
		ctx:           ctx,
		registry:      registry,
		username:      username,
		password:      password,
		client:        client,
		signatureBase: sigBase,
		scope: authScope{
			actions:    actions,
			remoteName: reference.Path(ref.ref),
		},
	}, nil
}

// makeRequest creates and executes a http.Request with the specified parameters, adding authentication and TLS options for the Docker client.
// The host name and schema is taken from the client or autodetected, and the path is relative to it, i.e. the path usually starts with /v2/.
func (c *dockerClient) makeRequest(ctx context.Context, method, path string, headers map[string][]string, stream io.Reader) (*http.Response, error) {
	if err := c.detectProperties(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s://%s%s", c.scheme, c.registry, path)
	return c.makeRequestToResolvedURL(ctx, method, url, headers, stream, -1, true)
}

// makeRequestToResolvedURL creates and executes a http.Request with the specified parameters, adding authentication and TLS options for the Docker client.
// streamLen, if not -1, specifies the length of the data expected on stream.
// makeRequest should generally be preferred.
// TODO(runcom): too many arguments here, use a struct
func (c *dockerClient) makeRequestToResolvedURL(ctx context.Context, method, url string, headers map[string][]string, stream io.Reader, streamLen int64, sendAuth bool) (*http.Response, error) {
	req, err := http.NewRequest(method, url, stream)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if streamLen != -1 { // Do not blindly overwrite if streamLen == -1, http.NewRequest above can figure out the length of bytes.Reader and similar objects without us having to compute it.
		req.ContentLength = streamLen
	}
	req.Header.Set("Docker-Distribution-API-Version", "registry/2.0")
	for n, h := range headers {
		for _, hh := range h {
			req.Header.Add(n, hh)
		}
	}
	if c.ctx != nil && c.ctx.DockerRegistryUserAgent != "" {
		req.Header.Add("User-Agent", c.ctx.DockerRegistryUserAgent)
	}
	if sendAuth {
		if err := c.setupRequestAuth(req); err != nil {
			return nil, err
		}
	}
	logrus.Debugf("%s %s", method, url)
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// we're using the challenges from the /v2/ ping response and not the one from the destination
// URL in this request because:
//
// 1) docker does that as well
// 2) gcr.io is sending 401 without a WWW-Authenticate header in the real request
//
// debugging: https://github.com/containers/image/pull/211#issuecomment-273426236 and follows up
func (c *dockerClient) setupRequestAuth(req *http.Request) error {
	if len(c.challenges) == 0 {
		return nil
	}
	schemeNames := make([]string, 0, len(c.challenges))
	for _, challenge := range c.challenges {
		schemeNames = append(schemeNames, challenge.Scheme)
		switch challenge.Scheme {
		case "basic":
			req.SetBasicAuth(c.username, c.password)
			return nil
		case "bearer":
			if c.token == nil || time.Now().After(c.tokenExpiration) {
				realm, ok := challenge.Parameters["realm"]
				if !ok {
					return errors.Errorf("missing realm in bearer auth challenge")
				}
				service, _ := challenge.Parameters["service"] // Will be "" if not present
				scope := fmt.Sprintf("repository:%s:%s", c.scope.remoteName, c.scope.actions)
				token, err := c.getBearerToken(req.Context(), realm, service, scope)
				if err != nil {
					return err
				}
				c.token = token
				c.tokenExpiration = token.IssuedAt.Add(time.Duration(token.ExpiresIn) * time.Second)
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.Token))
			return nil
		default:
			logrus.Debugf("no handler for %s authentication", challenge.Scheme)
		}
	}
	logrus.Infof("None of the challenges sent by server (%s) are supported, trying an unauthenticated request anyway", strings.Join(schemeNames, ", "))
	return nil
}

func (c *dockerClient) getBearerToken(ctx context.Context, realm, service, scope string) (*bearerToken, error) {
	authReq, err := http.NewRequest("GET", realm, nil)
	if err != nil {
		return nil, err
	}
	authReq = authReq.WithContext(ctx)
	getParams := authReq.URL.Query()
	if service != "" {
		getParams.Add("service", service)
	}
	if scope != "" {
		getParams.Add("scope", scope)
	}
	authReq.URL.RawQuery = getParams.Encode()
	if c.username != "" && c.password != "" {
		authReq.SetBasicAuth(c.username, c.password)
	}
	tr := newTransport()
	// TODO(runcom): insecure for now to contact the external token service
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: tr}
	res, err := client.Do(authReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusUnauthorized:
		return nil, errors.Errorf("unable to retrieve auth token: 401 unauthorized")
	case http.StatusOK:
		break
	default:
		return nil, errors.Errorf("unexpected http code: %d, URL: %s", res.StatusCode, authReq.URL)
	}
	tokenBlob, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var token bearerToken
	if err := json.Unmarshal(tokenBlob, &token); err != nil {
		return nil, err
	}
	if token.ExpiresIn < minimumTokenLifetimeSeconds {
		token.ExpiresIn = minimumTokenLifetimeSeconds
		logrus.Debugf("Increasing token expiration to: %d seconds", token.ExpiresIn)
	}
	if token.IssuedAt.IsZero() {
		token.IssuedAt = time.Now().UTC()
	}
	return &token, nil
}

func getAuth(ctx *types.SystemContext, registry string) (string, string, error) {
	if ctx != nil && ctx.DockerAuthConfig != nil {
		return ctx.DockerAuthConfig.Username, ctx.DockerAuthConfig.Password, nil
	}
	var dockerAuth dockerConfigFile
	dockerCfgPath := filepath.Join(getDefaultConfigDir(".docker"), dockerCfgFileName)
	if _, err := os.Stat(dockerCfgPath); err == nil {
		j, err := ioutil.ReadFile(dockerCfgPath)
		if err != nil {
			return "", "", err
		}
		if err := json.Unmarshal(j, &dockerAuth); err != nil {
			return "", "", err
		}

	} else if os.IsNotExist(err) {
		// try old config path
		oldDockerCfgPath := filepath.Join(getDefaultConfigDir(dockerCfgObsolete))
		if _, err := os.Stat(oldDockerCfgPath); err != nil {
			if os.IsNotExist(err) {
				return "", "", nil
			}
			return "", "", errors.Wrap(err, oldDockerCfgPath)
		}

		j, err := ioutil.ReadFile(oldDockerCfgPath)
		if err != nil {
			return "", "", err
		}
		if err := json.Unmarshal(j, &dockerAuth.AuthConfigs); err != nil {
			return "", "", err
		}

	} else if err != nil {
		return "", "", errors.Wrap(err, dockerCfgPath)
	}

	// I'm feeling lucky
	if c, exists := dockerAuth.AuthConfigs[registry]; exists {
		return decodeDockerAuth(c.Auth)
	}

	// bad luck; let's normalize the entries first
	registry = normalizeRegistry(registry)
	normalizedAuths := map[string]dockerAuthConfig{}
	for k, v := range dockerAuth.AuthConfigs {
		normalizedAuths[normalizeRegistry(k)] = v
	}
	if c, exists := normalizedAuths[registry]; exists {
		return decodeDockerAuth(c.Auth)
	}
	return "", "", nil
}

// detectProperties detects various properties of the registry.
// See the dockerClient documentation for members which are affected by this.
func (c *dockerClient) detectProperties(ctx context.Context) error {
	if c.scheme != "" {
		return nil
	}

	ping := func(scheme string) error {
		url := fmt.Sprintf(resolvedPingV2URL, scheme, c.registry)
		resp, err := c.makeRequestToResolvedURL(ctx, "GET", url, nil, nil, -1, true)
		logrus.Debugf("Ping %s err %#v", url, err)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		logrus.Debugf("Ping %s status %d", url, resp.StatusCode)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
			return errors.Errorf("error pinging repository, response code %d", resp.StatusCode)
		}
		c.challenges = parseAuthHeader(resp.Header)
		c.scheme = scheme
		c.supportsSignatures = resp.Header.Get("X-Registry-Supports-Signatures") == "1"
		return nil
	}
	err := ping("https")
	if err != nil && c.ctx != nil && c.ctx.DockerInsecureSkipTLSVerify {
		err = ping("http")
	}
	if err != nil {
		err = errors.Wrap(err, "pinging docker registry returned")
		if c.ctx != nil && c.ctx.DockerDisableV1Ping {
			return err
		}
		// best effort to understand if we're talking to a V1 registry
		pingV1 := func(scheme string) bool {
			url := fmt.Sprintf(resolvedPingV1URL, scheme, c.registry)
			resp, err := c.makeRequestToResolvedURL(ctx, "GET", url, nil, nil, -1, true)
			logrus.Debugf("Ping %s err %#v", url, err)
			if err != nil {
				return false
			}
			defer resp.Body.Close()
			logrus.Debugf("Ping %s status %d", url, resp.StatusCode)
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
				return false
			}
			return true
		}
		isV1 := pingV1("https")
		if !isV1 && c.ctx != nil && c.ctx.DockerInsecureSkipTLSVerify {
			isV1 = pingV1("http")
		}
		if isV1 {
			err = ErrV1NotSupported
		}
	}
	return err
}

// getExtensionsSignatures returns signatures from the X-Registry-Supports-Signatures API extension,
// using the original data structures.
func (c *dockerClient) getExtensionsSignatures(ctx context.Context, ref dockerReference, manifestDigest digest.Digest) (*extensionSignatureList, error) {
	path := fmt.Sprintf(extensionsSignaturePath, reference.Path(ref.ref), manifestDigest)
	res, err := c.makeRequest(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, client.HandleErrorResponse(res)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var parsedBody extensionSignatureList
	if err := json.Unmarshal(body, &parsedBody); err != nil {
		return nil, errors.Wrapf(err, "Error decoding signature list")
	}
	return &parsedBody, nil
}

func getDefaultConfigDir(confPath string) string {
	return filepath.Join(homedir.Get(), confPath)
}

type dockerAuthConfig struct {
	Auth string `json:"auth,omitempty"`
}

type dockerConfigFile struct {
	AuthConfigs map[string]dockerAuthConfig `json:"auths"`
}

func decodeDockerAuth(s string) (string, string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		// if it's invalid just skip, as docker does
		return "", "", nil
	}
	user := parts[0]
	password := strings.Trim(parts[1], "\x00")
	return user, password, nil
}

// convertToHostname converts a registry url which has http|https prepended
// to just an hostname.
// Copied from github.com/docker/docker/registry/auth.go
func convertToHostname(url string) string {
	stripped := url
	if strings.HasPrefix(url, "http://") {
		stripped = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		stripped = strings.TrimPrefix(url, "https://")
	}

	nameParts := strings.SplitN(stripped, "/", 2)

	return nameParts[0]
}

func normalizeRegistry(registry string) string {
	normalized := convertToHostname(registry)
	switch normalized {
	case "registry-1.docker.io", "docker.io":
		return "index.docker.io"
	}
	return normalized
}
