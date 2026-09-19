/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gvisor

import (
	"archive/tar"
	"compress/bzip2"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"k8s.io/minikube/pkg/libmachine/mcnutils"
)

const (
	nodeDir              = "/node"
	containerdConfigPath = "/etc/containerd/config.toml"
	// containerdConfigBackupPath lives next to the config on persistent disk,
	// so a reboot (which clears /tmp) can't orphan it.
	containerdConfigBackupPath = "/etc/containerd/config.toml.gvisor-bak"
	// legacyContainerdConfigBackupPath is where older releases stored the
	// backup. Disable() still honors it so upgraded clusters roll back cleanly.
	legacyContainerdConfigBackupPath = "/tmp/containerd-config.toml.bak"

	configFragment = `
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runsc]
  runtime_type = "io.containerd.runsc.v1"
  pod_annotations = [ "dev.gvisor.*" ]
`
)

// gvisorTarballURL points at the bz2 release archive. bz2 (not zstd) is
// deliberate: the stdlib decompresses it with no new dependencies, and the
// addon image has no zstd binary to shell out to.
func gvisorTarballURL() (string, error) {
	base, err := releaseURL()
	if err != nil {
		return "", err
	}
	return base + "gvisor.tar.bz2", nil
}

func releaseURL() (string, error) {
	return releaseArchURL(runtime.GOARCH)
}

func releaseArchURL(goarch string) (string, error) {
	arch := goarch
	switch arch {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	default:
		return "", fmt.Errorf("unsupported architecture %q: gvisor only ships x86_64 and aarch64 releases", goarch)
	}
	return fmt.Sprintf("https://storage.googleapis.com/gvisor/releases/release/latest/%s/", arch), nil
}

// Enable follows these steps for enabling gvisor in minikube:
//  1. creates necessary directories for storing binaries and runsc logs
//  2. downloads runsc and gvisor-containerd-shim
//  3. configures containerd
//  4. restarts containerd
func Enable() error {
	if err := makeGvisorDirs(); err != nil {
		return fmt.Errorf("creating directories on node: %w", err)
	}
	if err := downloadBinaries(); err != nil {
		return fmt.Errorf("downloading binaries: %w", err)
	}
	if err := configure(); err != nil {
		return fmt.Errorf("copying config files: %w", err)
	}
	if err := restartContainerd(); err != nil {
		return fmt.Errorf("restarting containerd: %w", err)
	}
	// When pod is terminated, disable gvisor and exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if err := Disable(); err != nil {
			log.Printf("Error disabling gvisor: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	log.Print("gvisor successfully enabled in cluster")
	// sleep for one year so the pod continuously runs
	select {}
}

// makeGvisorDirs creates necessary directories on the node
func makeGvisorDirs() error {
	// Make /run/containerd/runsc to hold logs
	fp := filepath.Join(nodeDir, "run/containerd/runsc")
	if err := os.MkdirAll(fp, 0755); err != nil {
		return fmt.Errorf("creating runsc dir: %w", err)
	}

	// Make /tmp/runsc to also hold logs
	fp = filepath.Join(nodeDir, "tmp/runsc")
	if err := os.MkdirAll(fp, 0755); err != nil {
		return fmt.Errorf("creating runsc logs dir: %w", err)
	}

	return nil
}

func downloadBinaries() error {
	url, err := gvisorTarballURL()
	if err != nil {
		return err
	}
	return downloadBinariesFrom(url, filepath.Join(nodeDir, "usr/bin"))
}

func downloadBinariesFrom(url, binDir string) error {
	tmp, err := downloadToTemp(url)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", url, err)
	}
	defer os.Remove(tmp)
	if err := extractGvisorTarball(tmp, binDir); err != nil {
		return fmt.Errorf("installing gvisor binaries: %w", err)
	}
	return nil
}

// downloadToTemp fetches url to a temp file, failing closed on non-200
// so an error page is never mistaken for a usable binary.
func downloadToTemp(url string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request for %s: %w", url, err)
	}
	req.Header.Set("User-Agent", "minikube")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}
	tmp, err := os.CreateTemp("", "gvisor-*.tar.bz2")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()
	n, copyErr := io.Copy(tmp, resp.Body)
	closeErr := tmp.Close()
	if copyErr != nil {
		os.Remove(tmpName)
		return "", fmt.Errorf("copying response body: %w", copyErr)
	}
	if closeErr != nil {
		os.Remove(tmpName)
		return "", fmt.Errorf("closing temp file: %w", closeErr)
	}
	if n == 0 {
		os.Remove(tmpName)
		return "", fmt.Errorf("empty response for %s", url)
	}
	return tmpName, nil
}

// extractGvisorTarball installs runsc, containerd-shim-runsc-v1 and the
// gvisor-bin sidecars runsc resolves next to its own binary.
func extractGvisorTarball(tarballPath, binDir string) error {
	f, err := os.Open(tarballPath)
	if err != nil {
		return fmt.Errorf("opening tarball: %w", err)
	}
	defer f.Close()
	return extractTarBz2(f, binDir)
}

func extractTarBz2(r io.Reader, binDir string) error {
	return extractTar(bzip2.NewReader(r), binDir)
}

func extractTar(r io.Reader, binDir string) error {
	tr := tar.NewReader(r)
	found := map[string]bool{"runsc": false, "containerd-shim-runsc-v1": false}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tarball: %w", err)
		}
		name := path.Clean(hdr.Name)
		name = strings.TrimPrefix(name, "./")
		name = strings.TrimPrefix(name, "/")
		if name == "." || name == "" || hasDotDotComponent(name) {
			return fmt.Errorf("rejecting unsafe tar entry %q", hdr.Name)
		}
		if !allowedTarEntry(name) {
			continue
		}
		if hdr.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(filepath.Join(binDir, name), 0755); err != nil {
				return fmt.Errorf("creating dir for %s: %w", name, err)
			}
			continue
		}
		if hdr.Typeflag == tar.TypeSymlink || hdr.Typeflag == tar.TypeLink {
			return fmt.Errorf("rejecting link tar entry %q", hdr.Name)
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}
		dest := filepath.Join(binDir, name)
		if !strings.HasPrefix(dest, filepath.Clean(binDir)+string(os.PathSeparator)) && dest != filepath.Clean(binDir) {
			return fmt.Errorf("rejecting tar entry outside bin dir %q", hdr.Name)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("creating dir for %s: %w", name, err)
		}
		if err := writeExecutable(tr, dest); err != nil {
			return fmt.Errorf("installing %s: %w", name, err)
		}
		if _, ok := found[name]; ok {
			found[name] = true
		}
	}
	for name, ok := range found {
		if !ok {
			return fmt.Errorf("tarball missing required binary %q", name)
		}
	}
	return nil
}

func hasDotDotComponent(name string) bool {
	for _, part := range strings.Split(name, "/") {
		if part == ".." {
			return true
		}
	}
	return false
}

func allowedTarEntry(name string) bool {
	if name == "runsc" || name == "containerd-shim-runsc-v1" {
		return true
	}
	if strings.HasPrefix(name, "gvisor-bin/") {
		base := strings.TrimPrefix(name, "gvisor-bin/")
		return base != "" && !strings.Contains(base, "/")
	}
	return false
}

func writeExecutable(src io.Reader, dest string) error {
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()
	n, copyErr := io.Copy(tmp, src)
	closeErr := tmp.Close()
	if copyErr != nil {
		os.Remove(tmpName)
		return fmt.Errorf("copying binary: %w", copyErr)
	}
	if closeErr != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", closeErr)
	}
	if n == 0 {
		os.Remove(tmpName)
		return fmt.Errorf("empty binary %q", dest)
	}
	if err := os.Chmod(tmpName, 0755); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("fixing perms: %w", err)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("moving %s into place: %w", dest, err)
	}
	return nil
}

// stanzaMarker identifies the runsc runtime block configure() manages, so a
// second Enable stays idempotent instead of appending a duplicate table.
const stanzaMarker = "runtimes.runsc"

// configure changes containerd `config.toml` file to include runsc runtime. A
// copy of the original file is stored next to it to be restored when this
// plug in is disabled.
func configure() error {
	return configureAt(nodeDir)
}

// configureAt appends the runsc stanza to the containerd config under root,
// keeping a pristine backup for Disable. It is safe to run repeatedly: when
// the stanza is already present (ours or the user's own runsc block) the
// config is left untouched, and an existing backup is never overwritten with
// patched content.
func configureAt(root string) error {
	configPath := filepath.Join(root, containerdConfigPath)
	backupPath := filepath.Join(root, containerdConfigBackupPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", configPath, err)
	}
	if hasGvisorStanza(string(data)) {
		log.Print("runsc stanza already present, skipping containerd config change")
		return nil
	}
	// Back up once: never overwrite an existing backup with patched content.
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		log.Printf("Storing default config.toml at %s", containerdConfigBackupPath)
		if err := mcnutils.CopyFile(configPath, backupPath); err != nil {
			return fmt.Errorf("copying default config.toml: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("checking backup %s: %w", backupPath, err)
	}

	// Append runsc configuration to containerd config.
	config, err := os.OpenFile(configPath, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return fmt.Errorf("opening %s: %w", configPath, err)
	}
	if _, err := config.WriteString(configFragment); err != nil {
		config.Close()
		return fmt.Errorf("changing config.toml: %w", err)
	}
	if err := config.Sync(); err != nil {
		config.Close()
		return fmt.Errorf("syncing config.toml: %w", err)
	}
	if err := config.Close(); err != nil {
		return fmt.Errorf("closing config.toml: %w", err)
	}
	return nil
}

// hasGvisorStanza reports whether a runsc runtime block is already
// configured. It matches the "runtimes.runsc" table name as a substring, so a
// user-supplied runsc block also counts: appending a second table would
// produce invalid TOML, so the user's block wins and ours is skipped.
func hasGvisorStanza(content string) bool {
	return strings.Contains(content, stanzaMarker)
}

func restartContainerd() error {
	log.Print("restartContainerd black magic happening")

	log.Print("Stopping rpc-statd.service...")
	cmd := exec.Command("/usr/sbin/chroot", "/node", "sudo", "systemctl", "stop", "rpc-statd.service")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return fmt.Errorf("stopping rpc-statd.service: %w", err)
	}

	log.Print("Restarting containerd...")
	cmd = exec.Command("/usr/sbin/chroot", "/node", "sudo", "systemctl", "restart", "containerd")
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Print(string(out))
		return fmt.Errorf("restarting containerd: %w", err)
	}

	log.Print("Starting rpc-statd...")
	cmd = exec.Command("/usr/sbin/chroot", "/node", "sudo", "systemctl", "start", "rpc-statd.service")
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Print(string(out))
		return fmt.Errorf("restarting rpc-statd.service: %w", err)
	}
	log.Print("containerd restart complete")
	return nil
}
