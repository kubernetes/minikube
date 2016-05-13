package persist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/drivers/none"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/hosttest"
)

func cleanup() {
	os.RemoveAll(os.Getenv("MACHINE_STORAGE_PATH"))
}

func getTestStore() Filestore {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	mcndirs.BaseDir = tmpDir

	return Filestore{
		Path:             tmpDir,
		CaCertPath:       filepath.Join(tmpDir, "certs", "ca-cert.pem"),
		CaPrivateKeyPath: filepath.Join(tmpDir, "certs", "ca-key.pem"),
	}
}

func TestStoreSave(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.GetMachinesDir(), h.Name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}

	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		r, err := regexp.Compile("config.json.tmp*")
		if err != nil {
			t.Fatalf("Failed to compile regexp string")
		}
		if r.MatchString(f.Name()) {
			t.Fatalf("Failed to remove temp filestore:%s", f.Name())
		}
	}
}

func TestStoreSaveOmitRawDriver(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	configJSONPath := filepath.Join(store.GetMachinesDir(), h.Name, "config.json")

	f, err := os.Open(configJSONPath)
	if err != nil {
		t.Fatal(err)
	}

	configData, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	fakeHost := make(map[string]interface{})

	if err := json.Unmarshal(configData, &fakeHost); err != nil {
		t.Fatal(err)
	}

	if rawDriver, ok := fakeHost["RawDriver"]; ok {
		t.Fatal("Should not have gotten a value for RawDriver reading host from disk but got one: ", rawDriver)
	}

}

func TestStoreRemove(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.GetMachinesDir(), h.Name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}

	err = store.Remove(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); err == nil {
		t.Fatalf("Host path still exists after remove: %s", path)
	}
}

func TestStoreList(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	hosts, err := store.List()
	if len(hosts) != 1 {
		t.Fatalf("List returned %d items, expected 1", len(hosts))
	}

	if hosts[0] != h.Name {
		t.Fatalf("hosts[0] name is incorrect, got: %s", hosts[0])
	}
}

func TestStoreExists(t *testing.T) {
	defer cleanup()
	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	exists, err := store.Exists(h.Name)
	if exists {
		t.Fatal("Host should not exist before saving")
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	exists, err = store.Exists(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("Host should exist after saving")
	}

	if err := store.Remove(h.Name); err != nil {
		t.Fatal(err)
	}

	exists, err = store.Exists(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		t.Fatal("Host should not exist after removing")
	}
}

func TestStoreLoad(t *testing.T) {
	defer cleanup()

	expectedURL := "unix:///foo/baz"
	flags := hosttest.GetTestDriverFlags()
	flags.Data["url"] = expectedURL

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := h.Driver.SetConfigFromFlags(flags); err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	h, err = store.Load(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	rawDataDriver, ok := h.Driver.(*host.RawDataDriver)
	if !ok {
		t.Fatal("Expected driver loaded from store to be of type *host.RawDataDriver and it was not")
	}

	realDriver := none.NewDriver(h.Name, store.Path)

	if err := json.Unmarshal(rawDataDriver.Data, &realDriver); err != nil {
		t.Fatalf("Error unmarshaling rawDataDriver data into concrete 'none' driver: %s", err)
	}

	h.Driver = realDriver

	actualURL, err := h.URL()
	if err != nil {
		t.Fatal(err)
	}

	if actualURL != expectedURL {
		t.Fatalf("GetURL is not %q, got %q", expectedURL, actualURL)
	}
}
