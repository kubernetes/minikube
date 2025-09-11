package sshutil

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// GenerateSSHKey generates an ed25519 SSH key pair at the provided path.
// The public key is written to the same location with a ".pub" suffix.
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

	pubKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		return fmt.Errorf("error creating public key: %w", err)
	}

	block, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return fmt.Errorf("error marshalling private key: %w", err)
	}

	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0o600); err != nil {
		return fmt.Errorf("error writing private key: %w", err)
	}

	pubPath := fmt.Sprintf("%s.pub", path)
	if err := os.WriteFile(pubPath, ssh.MarshalAuthorizedKey(pubKey), 0o644); err != nil {
		return fmt.Errorf("error writing public key: %w", err)
	}

	return nil
}
