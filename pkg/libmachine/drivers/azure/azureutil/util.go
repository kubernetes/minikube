package azureutil

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

/* Utilities */

// randomAzureStorageAccountName generates a valid storage account name prefixed
// with a predefined string. Availability of the name is not checked. Uses maximum
// length to maximise randomness.
func randomAzureStorageAccountName() string {
	const (
		maxLen = 24
		chars  = "0123456789abcdefghijklmnopqrstuvwxyz"
	)
	return storageAccountPrefix + randomString(maxLen-len(storageAccountPrefix), chars)
}

// randomString generates a random string of given length using specified alphabet.
func randomString(n int, alphabet string) string {
	r := timeSeed()
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(b)
}

// imageName holds various components of an OS image name identifier
type imageName struct{ publisher, offer, sku, version string }

// parseImageName parses a publisher:offer:sku:version into those parts.
func parseImageName(image string) (imageName, error) {
	l := strings.Split(image, ":")
	if len(l) != 4 {
		return imageName{}, fmt.Errorf("image name %q not a valid format", image)
	}
	return imageName{l[0], l[1], l[2], l[3]}, nil
}

func timeSeed() *rand.Rand { return rand.New(rand.NewSource(time.Now().UTC().UnixNano())) }
