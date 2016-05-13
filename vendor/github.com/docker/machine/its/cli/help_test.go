package cli

import (
	"testing"

	"github.com/docker/machine/its"
)

func TestHelp(t *testing.T) {
	test := its.NewTest(t)
	defer test.TearDown()

	test.Run("cli: show info", func() {
		test.Machine("").Should().Succeed("Usage:", "Create and manage machines running Docker")
	})

	test.Run("cli: show active help", func() {
		test.Machine("active -h").Should().Succeed("machine active")
	})

	test.Run("cli: show config help", func() {
		test.Machine("config -h").Should().Succeed("machine config")
	})

	test.Run("cli: show create help", func() {
		test.Machine("create -h").Should().Succeed("machine create")
	})

	test.Run("cli: show env help", func() {
		test.Machine("env -h").Should().Succeed("machine env")
	})

	test.Run("cli: show inspect help", func() {
		test.Machine("inspect -h").Should().Succeed("machine inspect")
	})

	test.Run("cli: show ip help", func() {
		test.Machine("ip -h").Should().Succeed("machine ip")
	})

	test.Run("cli: show kill help", func() {
		test.Machine("kill -h").Should().Succeed("machine kill")
	})

	test.Run("cli: show ls help", func() {
		test.Machine("ls -h").Should().Succeed("machine ls")
	})

	test.Run("cli: show regenerate-certs help", func() {
		test.Machine("regenerate-certs -h").Should().Succeed("machine regenerate-certs")
	})

	test.Run("cli: show restart help", func() {
		test.Machine("restart -h").Should().Succeed("machine restart")
	})

	test.Run("cli: show rm help", func() {
		test.Machine("rm -h").Should().Succeed("machine rm")
	})

	test.Run("cli: show scp help", func() {
		test.Machine("scp -h").Should().Succeed("machine scp")
	})

	test.Run("cli: show ssh help", func() {
		test.Machine("ssh -h").Should().Succeed("machine ssh")
	})

	test.Run("cli: show start help", func() {
		test.Machine("start -h").Should().Succeed("machine start")
	})

	test.Run("cli: show status help", func() {
		test.Machine("status -h").Should().Succeed("machine status")
	})

	test.Run("cli: show stop help", func() {
		test.Machine("stop -h").Should().Succeed("machine stop")
	})

	test.Run("cli: show upgrade help", func() {
		test.Machine("upgrade -h").Should().Succeed("machine upgrade")
	})

	test.Run("cli: show url help", func() {
		test.Machine("url -h").Should().Succeed("machine url")
	})

	test.Run("cli: show version", func() {
		test.Machine("-v").Should().Succeed("version")
	})

	test.Run("cli: show help", func() {
		test.Machine("--help").Should().Succeed("Usage:")
	})
}
