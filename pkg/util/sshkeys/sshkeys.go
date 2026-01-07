/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package sshkeys

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"runtime"

	"github.com/hectane/go-acl"
	"golang.org/x/crypto/ssh"
)

const (
	Ed25519KeyName = "id_ed25519"
	RSAKeyName     = "id_rsa"
)

// ResolveKeyPath returns the first existing key path, or primary when neither exists.
func ResolveKeyPath(primary, fallback string) string {
	if fileExists(primary) {
		return primary
	}
	if fallback != "" && fileExists(fallback) {
		return fallback
	}
	return primary
}

// GenerateSSHKey creates an ed25519 SSH keypair at the given private key path.
func GenerateSSHKey(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("desired directory for SSH keys does not exist: %w", err)
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("error generating key pair: %w", err)
	}

	block, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return fmt.Errorf("error marshaling private key: %w", err)
	}
	if err := writeKeyFile(path, pem.EncodeToMemory(block)); err != nil {
		return fmt.Errorf("error writing private key: %w", err)
	}

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return fmt.Errorf("error converting public key: %w", err)
	}
	if err := writeKeyFile(path+".pub", ssh.MarshalAuthorizedKey(sshPub)); err != nil {
		return fmt.Errorf("error writing public key: %w", err)
	}

	return nil
}

func writeKeyFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return chmodFile(path, 0600)
}

func chmodFile(path string, mode os.FileMode) error {
	switch runtime.GOOS {
	case "darwin", "freebsd", "linux", "openbsd":
		return os.Chmod(path, mode)
	case "windows":
		return acl.Chmod(path, mode)
	default:
		return os.Chmod(path, mode)
	}
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
