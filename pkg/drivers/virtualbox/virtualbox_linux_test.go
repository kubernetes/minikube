package virtualbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShareName(t *testing.T) {
	name, dir := getShareDriveAndName()

	assert.Equal(t, name, "hosthome")
	assert.Equal(t, dir, "/home")
}
