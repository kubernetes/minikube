package cli

import (
	"testing"

	"github.com/docker/machine/its"
)

func TestCreateRm(t *testing.T) {
	test := its.NewTest(t)
	defer test.TearDown()

	test.Run("non-existent driver fails", func() {
		test.Machine("create -d bogus bogus").Should().Fail(`Driver "bogus" not found. Do you have the plugin binary accessible in your PATH?`)
	})

	test.Run("non-existent driver fails", func() {
		test.Machine("create -d bogus bogus").Should().Fail(`Driver "bogus" not found. Do you have the plugin binary accessible in your PATH?`)
	})

	test.Run("create with no name fails", func() {
		test.Machine("create -d none").Should().Fail(`Error: No machine name specified`)
	})

	test.Run("create with invalid name fails", func() {
		test.Machine("create -d none --url none ∞").Should().Fail(`Error creating machine: Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -`)
	})

	test.Run("create with invalid name fails", func() {
		test.Machine("create -d none --url none -").Should().Fail(`Error creating machine: Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -`)
	})

	test.Run("create with invalid name fails", func() {
		test.Machine("create -d none --url none .").Should().Fail(`Error creating machine: Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -`)
	})

	test.Run("create with invalid name fails", func() {
		test.Machine("create -d none --url none ..").Should().Fail(`Error creating machine: Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -`)
	})

	test.Run("create with weird but valid name succeeds", func() {
		test.Machine("create -d none --url none a").Should().Succeed()
	})

	test.Run("fail with extra argument", func() {
		test.Machine("create -d none --url none a extra").Should().Fail(`Invalid command line. Found extra arguments [extra]`)
	})

	test.Run("create with weird but valid name", func() {
		test.Machine("create -d none --url none 0").Should().Succeed()
	})

	test.Run("rm with no name fails", func() {
		test.Machine("rm -y").Should().Fail(`Error: Expected to get one or more machine names as arguments`)
	})

	test.Run("rm non existent machine fails", func() {
		test.Machine("rm ∞ -y").Should().Fail(`Error removing host "∞": Host does not exist: "∞"`)
	})

	test.Run("rm existing machine", func() {
		test.Machine("rm 0 -y").Should().Succeed()
	})

	test.Run("rm ask user confirmation when -y is not provided", func() {
		test.Machine("create -d none --url none ba").Should().Succeed()
		test.Cmd("echo y | machine rm ba").Should().Succeed()
	})

	test.Run("rm deny user confirmation when -y is not provided", func() {
		test.Machine("create -d none --url none ab").Should().Succeed()
		test.Cmd("echo n | machine rm ab").Should().Succeed()
	})

	test.Run("rm never prompt user confirmation when -f is provided", func() {
		test.Machine("create -d none --url none c").Should().Succeed()
		test.Machine("rm -f c").Should().Succeed("Successfully removed c")
	})
}
