package cli

import (
	"testing"

	"github.com/docker/machine/its"
)

func TestDriverHelp(t *testing.T) {
	test := its.NewTest(t)
	defer test.TearDown()

	test.SkipDriver("ci-test")

	test.Run("no --help flag or command specified", func() {
		test.Machine("create -d $DRIVER").Should().Fail("Error: No machine name specified")
	})

	test.Run("-h flag specified", func() {
		test.Machine("create -d $DRIVER -h").Should().Succeed(test.DriverName())
	})

	test.Run("--help flag specified", func() {
		test.Machine("create -d $DRIVER --help").Should().Succeed(test.DriverName())
	})
}
