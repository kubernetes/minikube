package b2d

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/docker/machine/libmachine/log"
)

const (
	defaultURL            = "https://api.github.com/repos/boot2docker/boot2docker/releases/latest"
	defaultISOFilename    = "boot2docker.iso"
	defaultVolumeIDOffset = int64(0x8028)
	defaultVolumeIDLength = 32
)

var (
	GithubAPIToken string
)

var (
	errGitHubAPIResponse = errors.New(`Error getting a version tag from the Github API response.
You may be getting rate limited by Github.`)
)

var (
	AUFSBugB2DVersions = map[string]string{
		"v1.9.1": "https://github.com/docker/docker/issues/18180",
	}
)

func defaultTimeout(network, addr string) (net.Conn, error) {
	return net.Dial(network, addr)
}

func getClient() *http.Client {
	transport := http.Transport{
		DisableKeepAlives: true,
		Proxy:             http.ProxyFromEnvironment,
		Dial:              defaultTimeout,
	}

	return &http.Client{
		Transport: &transport,
	}
}

func getRequest(apiURL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	if GithubAPIToken != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", GithubAPIToken))
	}

	return req, nil
}

// releaseGetter is a client that gets release information of a product and downloads it.
type releaseGetter interface {
	// Filename returns filename of the product.
	Filename() string
	// getReleaseTag gets a release tag from the given URL.
	getReleaseTag(apiURL string) (string, error)
	// GetReleaseURL gets the latest release download URL from the given URL.
	GetReleaseURL(apiURL string) (string, error)
}

// b2dReleaseGetter implements the releaseGetter interface for getting the release of Boot2Docker.
type b2dReleaseGetter struct {
	isoFilename string
}

func (b *b2dReleaseGetter) Filename() string {
	if b == nil {
		return ""
	}
	return b.isoFilename
}

// getReleaseTag gets the release tag of Boot2Docker from apiURL.
func (*b2dReleaseGetter) getReleaseTag(apiURL string) (string, error) {
	if apiURL == "" {
		apiURL = defaultURL
	}

	client := getClient()
	req, err := getRequest(apiURL)
	if err != nil {
		return "", err
	}
	rsp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	var t struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(rsp.Body).Decode(&t); err != nil {
		return "", err
	}
	if t.TagName == "" {
		return "", errGitHubAPIResponse
	}
	return t.TagName, nil
}

// GetReleaseURL gets the latest release URL of Boot2Docker.
// FIXME: find or create some other way to get the "latest release" of boot2docker since the GitHub API has a pretty low rate limit on API requests
func (b *b2dReleaseGetter) GetReleaseURL(apiURL string) (string, error) {
	if apiURL == "" {
		apiURL = defaultURL
	}

	// match github (enterprise) release urls:
	// https://api.github.com/repos/../../releases/latest or
	// https://some.github.enterprise/api/v3/repos/../../releases/latest
	re := regexp.MustCompile("(https?)://([^/]+)(/api/v3)?/repos/([^/]+)/([^/]+)/releases/latest")
	matches := re.FindStringSubmatch(apiURL)
	if len(matches) != 6 {
		// does not match a github releases api URL
		return apiURL, nil
	}

	scheme, host, org, repo := matches[1], matches[2], matches[4], matches[5]
	if host == "api.github.com" {
		host = "github.com"
	}

	tag, err := b.getReleaseTag(apiURL)
	if err != nil {
		return "", err
	}

	log.Infof("Latest release for %s/%s/%s is %s", host, org, repo, tag)
	bugURL, ok := AUFSBugB2DVersions[tag]
	if ok {
		log.Warnf(`
Boot2Docker %s has a known issue with AUFS.
See here for more details: %s
Consider specifying another storage driver (e.g. 'overlay') using '--engine-storage-driver' instead.
`, tag, bugURL)
	}
	url := fmt.Sprintf("%s://%s/%s/%s/releases/download/%s/%s", scheme, host, org, repo, tag, b.isoFilename)
	return url, nil
}

// iso is an ISO volume.
type iso interface {
	// path returns the path of the ISO.
	path() string
	// exists reports whether the ISO exists.
	Exists() bool
	// version returns version information of the ISO.
	version() (string, error)
}

// b2dISO represents a Boot2Docker ISO. It implements the ISO interface.
type b2dISO struct {
	// path of Boot2Docker ISO
	commonIsoPath string

	// offset and length of ISO volume ID
	// cf. http://serverfault.com/questions/361474/is-there-a-way-to-change-a-iso-files-volume-id-from-the-command-line
	volumeIDOffset int64
	volumeIDLength int
}

func (b *b2dISO) path() string {
	if b == nil {
		return ""
	}
	return b.commonIsoPath
}

func (b *b2dISO) Exists() bool {
	if b == nil {
		return false
	}

	_, err := os.Stat(b.commonIsoPath)
	return !os.IsNotExist(err)
}

// version scans the volume ID in b and returns its version tag.
func (b *b2dISO) version() (string, error) {
	if b == nil {
		return "", nil
	}

	iso, err := os.Open(b.commonIsoPath)
	if err != nil {
		return "", err
	}
	defer iso.Close()

	isoMetadata := make([]byte, b.volumeIDLength)
	_, err = iso.ReadAt(isoMetadata, b.volumeIDOffset)
	if err != nil {
		return "", err
	}

	verRegex := regexp.MustCompile(`v\d+\.\d+\.\d+`)
	ver := string(verRegex.Find(isoMetadata))
	log.Debug("local Boot2Docker ISO version: ", ver)
	return ver, nil
}

type B2dUtils struct {
	releaseGetter
	iso
	storePath    string
	ImgCachePath string
}

func NewB2dUtils(storePath string) *B2dUtils {
	imgCachePath := filepath.Join(storePath, "cache")

	return &B2dUtils{
		releaseGetter: &b2dReleaseGetter{isoFilename: defaultISOFilename},
		iso: &b2dISO{
			commonIsoPath:  filepath.Join(imgCachePath, defaultISOFilename),
			volumeIDOffset: defaultVolumeIDOffset,
			volumeIDLength: defaultVolumeIDLength,
		},
		storePath:    storePath,
		ImgCachePath: imgCachePath,
	}
}

func (b *B2dUtils) IsLatest() bool {
	localVer, err := b.version()
	if err != nil {
		log.Warn("Unable to get the local Boot2Docker ISO version: ", err)
		return false
	}

	latestVer, err := b.getReleaseTag("")
	if err != nil {
		log.Warn("Unable to get the latest Boot2Docker ISO release version: ", err)
		return true
	}

	return localVer == latestVer
}
