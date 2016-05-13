package cli

import (
	"testing"

	"github.com/docker/machine/its"
)

func TestInspect(t *testing.T) {
	test := its.NewTest(t)
	defer test.TearDown()

	test.Run("inspect: show error in case of no args", func() {
		test.Machine("inspect").Should().Fail(`Error: No machine name(s) specified and no "default" machine exists.`)
	})
}
