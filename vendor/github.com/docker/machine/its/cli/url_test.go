package cli

import (
	"testing"

	"github.com/docker/machine/its"
)

func TestUrl(t *testing.T) {
	test := its.NewTest(t)
	defer test.TearDown()

	test.Run("url: show error in case of no args", func() {
		test.Machine("url").Should().Fail(`Error: No machine name(s) specified and no "default" machine exists.`)
	})
}
