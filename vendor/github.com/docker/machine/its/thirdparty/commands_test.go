package thirdparty

import (
	"testing"

	"github.com/docker/machine/its"
)

func TestThirdPartyCompatibility(t *testing.T) {
	test := its.NewTest(t)
	defer test.TearDown()

	test.RequireDriver("ci-test")

	test.Run("create", func() {
		test.Machine("create -d $DRIVER --url url default").Should().Succeed()
	})

	test.Run("ls", func() {
		test.Machine("ls -q").Should().Succeed().ContainLines(1).EqualLine(0, "default")
	})

	test.Run("url", func() {
		test.Machine("url default").Should().Succeed("url")
	})

	test.Run("status", func() {
		test.Machine("status default").Should().Succeed("Running")
	})

	test.Run("rm", func() {
		test.Machine("rm -y default").Should().Succeed()
	})
}
