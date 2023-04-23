package mcnutils

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/version"
	"github.com/stretchr/testify/assert"
)

func TestGetReleaseURL(t *testing.T) {
	testCases := []struct {
		apiURL         string
		isoURL         string
		machineVersion string
		response       string
	}{
		{"/repos/org/repo/releases/latest", "/org/repo/releases/download/v0.1/boot2docker.iso", "v0.7.0", `{"tag_name": "v0.1"}`},

		// Note the difference in this one: It's an RC version.
		{"/repos/org/repo/releases", "/org/repo/releases/download/v0.2-rc1/boot2docker.iso", "v0.7.0-rc2", `[{"tag_name": "v0.2-rc1"}, {"tag_name": "v0.1"}]`},

		{"http://dummy.com/boot2docker.iso", "http://dummy.com/boot2docker.iso", "v0.7.0", `{"tag_name": "v0.1"}`},
	}

	for _, tt := range testCases {
		testServer := newTestServer(tt.response)

		// TODO: Modifying this package level variable is not elegant,
		// but it is effective.  Ideally this should be exposed through
		// an interface.
		actualMachineVersion := version.Version
		version.Version = tt.machineVersion
		b := NewB2dUtils("/tmp/isos")
		isoURL, err := b.getReleaseURL(testServer.URL + tt.apiURL)

		assert.NoError(t, err)
		assert.Equal(t, testServer.URL+tt.isoURL, isoURL)
		version.Version = actualMachineVersion

		testServer.Close()
	}
}

func TestGetReleaseURLError(t *testing.T) {
	// GitHub API error response in case of rate limit
	ts := newTestServer(`{"message": "API rate limit exceeded for 127.0.0.1.",
		"documentation_url": "https://developer.github.com/v3/#rate-limiting"}`)
	defer ts.Close()

	testCases := []struct {
		apiURL string
	}{
		{ts.URL + "/repos/org/repo/releases/latest"},
		{"http://127.0.0.1/repos/org/repo/releases/latest"}, // dummy API URL. cannot connect it.
	}

	for _, tt := range testCases {
		b := NewB2dUtils("/tmp/isos")
		_, err := b.getReleaseURL(tt.apiURL)

		assert.Error(t, err)
	}
}

func TestVersion(t *testing.T) {
	testCases := []string{
		"v0.1.0",
		"v0.2.0-rc1",
	}

	for _, vers := range testCases {
		isopath, off, err := newDummyISO("", defaultISOFilename, vers)

		assert.NoError(t, err)

		b := &b2dISO{
			commonIsoPath:  isopath,
			volumeIDOffset: off,
			volumeIDLength: defaultVolumeIDLength,
		}

		got, err := b.version()

		assert.NoError(t, err)
		assert.Equal(t, vers, string(got))
		removeFileIfExists(isopath)
	}
}

func TestDownloadISO(t *testing.T) {
	testData := "test-download"
	ts := newTestServer(testData)
	defer ts.Close()

	filename := "test"

	tmpDir, err := ioutil.TempDir("", "machine-test-")

	assert.NoError(t, err)

	b := NewB2dUtils("/tmp/artifacts")
	err = b.DownloadISO(tmpDir, filename, ts.URL)

	assert.NoError(t, err)

	data, err := ioutil.ReadFile(filepath.Join(tmpDir, filename))

	assert.NoError(t, err)
	assert.Equal(t, testData, string(data))
}

func TestGetRequest(t *testing.T) {
	testCases := []struct {
		token string
		want  string
	}{
		{"", ""},
		{"CATBUG", "token CATBUG"},
	}

	for _, tt := range testCases {
		GithubAPIToken = tt.token

		req, err := getRequest("http://some.github.api")

		assert.NoError(t, err)
		assert.Equal(t, tt.want, req.Header.Get("Authorization"))
	}
}

type MockReadCloser struct {
	blockLengths []int
	currentBlock int
}

func (r *MockReadCloser) Read(p []byte) (n int, err error) {
	n = r.blockLengths[r.currentBlock]
	r.currentBlock++
	return
}

func (r *MockReadCloser) Close() error {
	return nil
}

func TestReaderWithProgress(t *testing.T) {
	readCloser := MockReadCloser{blockLengths: []int{5, 45, 50}}
	output := new(bytes.Buffer)
	buffer := make([]byte, 100)

	readerWithProgress := ReaderWithProgress{
		ReadCloser:     &readCloser,
		out:            output,
		expectedLength: 100,
	}

	readerWithProgress.Read(buffer)
	assert.Equal(t, "0%..", output.String())

	readerWithProgress.Read(buffer)
	assert.Equal(t, "0%....10%....20%....30%....40%....50%", output.String())

	readerWithProgress.Read(buffer)
	assert.Equal(t, "0%....10%....20%....30%....40%....50%....60%....70%....80%....90%....100%", output.String())

	readerWithProgress.Close()
	assert.Equal(t, "0%....10%....20%....30%....40%....50%....60%....70%....80%....90%....100%\n", output.String())
}

type mockReleaseGetter struct {
	ver    string
	apiErr error
	verCh  chan<- string
}

func (m *mockReleaseGetter) filename() string {
	return defaultISOFilename
}

func (m *mockReleaseGetter) getReleaseTag(apiURL string) (string, error) {
	return m.ver, m.apiErr
}

func (m *mockReleaseGetter) getReleaseURL(apiURL string) (string, error) {
	return "http://127.0.0.1/dummy", m.apiErr
}

func (m *mockReleaseGetter) download(dir, file, isoURL string) error {
	path := filepath.Join(dir, file)
	var err error
	if _, e := os.Stat(path); os.IsNotExist(e) {
		err = ioutil.WriteFile(path, dummyISOData("  ", m.ver), 0644)
	}

	// send a signal of downloading the latest version
	m.verCh <- m.ver
	return err
}

type mockISO struct {
	isopath string
	exist   bool
	ver     string
	verCh   <-chan string
}

func (m *mockISO) path() string {
	return m.isopath
}

func (m *mockISO) exists() bool {
	return m.exist
}

func (m *mockISO) version() (string, error) {
	select {
	// receive version of a downloaded iso
	case ver := <-m.verCh:
		return ver, nil
	default:
		return m.ver, nil
	}
}

func TestCopyDefaultISOToMachine(t *testing.T) {
	apiErr := errors.New("api error")

	testCases := []struct {
		machineName string
		create      bool
		localVer    string
		latestVer   string
		apiErr      error
		wantVer     string
	}{
		{"none", false, "", "v1.0.0", nil, "v1.0.0"},         // none => downloading
		{"latest", true, "v1.0.0", "v1.0.0", nil, "v1.0.0"},  // latest iso => as is
		{"old-badurl", true, "v0.1.0", "", apiErr, "v0.1.0"}, // old iso with bad api => as is
		{"old", true, "v0.1.0", "v1.0.0", nil, "v1.0.0"},     // old iso => updating
	}

	var isopath string
	var err error
	verCh := make(chan string, 1)
	for _, tt := range testCases {
		if tt.create {
			isopath, _, err = newDummyISO("cache", defaultISOFilename, tt.localVer)
		} else {
			if dir, e := ioutil.TempDir("", "machine-test"); e == nil {
				isopath = filepath.Join(dir, "cache", defaultISOFilename)
			}
		}

		// isopath: "$TMPDIR/machine-test-xxxxxx/cache/boot2docker.iso"
		// tmpDir: "$TMPDIR/machine-test-xxxxxx"
		imgCachePath := filepath.Dir(isopath)
		storePath := filepath.Dir(imgCachePath)

		b := &B2dUtils{
			releaseGetter: &mockReleaseGetter{
				ver:    tt.latestVer,
				apiErr: tt.apiErr,
				verCh:  verCh,
			},
			iso: &mockISO{
				isopath: isopath,
				exist:   tt.create,
				ver:     tt.localVer,
				verCh:   verCh,
			},
			storePath:    storePath,
			imgCachePath: imgCachePath,
		}

		dir := filepath.Join(storePath, "machines", tt.machineName)
		err = os.MkdirAll(dir, 0700)
		assert.NoError(t, err, "machine: %s", tt.machineName)

		err = b.CopyIsoToMachineDir("", tt.machineName)
		assert.NoError(t, err)

		dest := filepath.Join(dir, b.filename())
		_, pathErr := os.Stat(dest)

		assert.NoError(t, err, "machine: %s", tt.machineName)
		assert.True(t, !os.IsNotExist(pathErr), "machine: %s", tt.machineName)

		ver, err := b.version()

		assert.NoError(t, err, "machine: %s", tt.machineName)
		assert.Equal(t, tt.wantVer, ver, "machine: %s", tt.machineName)

		err = removeFileIfExists(isopath)
		assert.NoError(t, err, "machine: %s", tt.machineName)
	}
}

// newTestServer creates a new httptest.Server that returns respText as a response body.
func newTestServer(respText string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(respText))
	}))
}

// newDummyISO creates a dummy ISO file that contains the given version info,
// and returns its path and offset value to fetch the version info.
func newDummyISO(dir, name, version string) (string, int64, error) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		return "", 0, err
	}

	tmpDir = filepath.Join(tmpDir, dir)
	if e := os.MkdirAll(tmpDir, 755); e != nil {
		return "", 0, err
	}

	isopath := filepath.Join(tmpDir, name)
	log.Info("TEST: dummy ISO created at ", isopath)

	// dummy ISO data mimicking the real byte data of a Boot2Docker ISO image
	padding := "     "
	data := dummyISOData(padding, version)
	return isopath, int64(len(padding)), ioutil.WriteFile(isopath, data, 0644)
}

// dummyISOData returns mock data that contains given padding and version.
func dummyISOData(padding, version string) []byte {
	return []byte(fmt.Sprintf("%sBoot2Docker-%s                    ", padding, version))
}
