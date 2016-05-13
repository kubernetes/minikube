package cli

import (
	"testing"

	"github.com/docker/machine/its"
)

func TestLs(t *testing.T) {
	test := its.NewTest(t)
	defer test.TearDown()

	test.Run("setup", func() {
		test.Machine("create -d none --url url5 --engine-label app=1 testmachine5").Should().Succeed()
		test.Machine("create -d none --url url4 --engine-label foo=bar --engine-label app=1 testmachine4").Should().Succeed()
		test.Machine("create -d none --url url3 testmachine3").Should().Succeed()
		test.Machine("create -d none --url url2 testmachine2").Should().Succeed()
		test.Machine("create -d none --url url1 testmachine").Should().Succeed()
	})

	test.Run("ls: no filter", func() {
		test.Machine("ls").Should().Succeed().
			ContainLines(6).
			MatchLine(0, "NAME[ ]+ACTIVE[ ]+DRIVER[ ]+STATE[ ]+URL[ ]+SWARM[ ]+DOCKER[ ]+ERRORS").
			MatchLine(1, "testmachine[ ]+-[ ]+none[ ]+Running[ ]+url1[ ]+Unknown[ ]+Unable to query docker version: .*").
			MatchLine(2, "testmachine2[ ]+-[ ]+none[ ]+Running[ ]+url2[ ]+Unknown[ ]+Unable to query docker version: .*").
			MatchLine(3, "testmachine3[ ]+-[ ]+none[ ]+Running[ ]+url3[ ]+Unknown[ ]+Unable to query docker version: .*").
			MatchLine(4, "testmachine4[ ]+-[ ]+none[ ]+Running[ ]+url4[ ]+Unknown[ ]+Unable to query docker version: .*").
			MatchLine(5, "testmachine5[ ]+-[ ]+none[ ]+Running[ ]+url5[ ]+Unknown[ ]+Unable to query docker version: .*")
	})

	test.Run("ls: filter on label", func() {
		test.Machine("ls --filter label=foo=bar").Should().Succeed().
			ContainLines(2).
			ContainLine(0, "NAME").
			ContainLine(1, "testmachine4")
	})

	test.Run("ls: mutiple filters on label", func() {
		test.Machine("ls --filter label=foo=bar --filter label=app=1").Should().Succeed().
			ContainLines(3).
			ContainLine(0, "NAME").
			ContainLine(1, "testmachine4").
			ContainLine(2, "testmachine5")
	})

	test.Run("ls: non-existing filter on label", func() {
		test.Machine("ls --filter label=invalid=filter").Should().Succeed().
			ContainLines(1).
			ContainLine(0, "NAME")
	})

	test.Run("ls: filter on driver", func() {
		test.Machine("ls --filter driver=none").Should().Succeed().
			ContainLines(6).
			ContainLine(0, "NAME").
			ContainLine(1, "testmachine").
			ContainLine(2, "testmachine2").
			ContainLine(3, "testmachine3").
			ContainLine(4, "testmachine4").
			ContainLine(5, "testmachine5")
	})

	test.Run("ls: filter on driver", func() {
		test.Machine("ls -q --filter driver=none").Should().Succeed().
			ContainLines(5).
			EqualLine(0, "testmachine").
			EqualLine(1, "testmachine2").
			EqualLine(2, "testmachine3")
	})

	test.Run("ls: filter on state", func() {
		test.Machine("ls --filter state=Running").Should().Succeed().
			ContainLines(6).
			ContainLine(0, "NAME").
			ContainLine(1, "testmachine").
			ContainLine(2, "testmachine2").
			ContainLine(3, "testmachine3")

		test.Machine("ls -q --filter state=Running").Should().Succeed().
			ContainLines(5).
			EqualLine(0, "testmachine").
			EqualLine(1, "testmachine2").
			EqualLine(2, "testmachine3")

		test.Machine("ls --filter state=None").Should().Succeed().
			ContainLines(1).
			ContainLine(0, "NAME")

		test.Machine("ls --filter state=Paused").Should().Succeed().
			ContainLines(1).
			ContainLine(0, "NAME")

		test.Machine("ls --filter state=Saved").Should().Succeed().
			ContainLines(1).
			ContainLine(0, "NAME")

		test.Machine("ls --filter state=Stopped").Should().Succeed().
			ContainLines(1).
			ContainLine(0, "NAME")

		test.Machine("ls --filter state=Stopping").Should().Succeed().
			ContainLines(1).
			ContainLine(0, "NAME")

		test.Machine("ls --filter state=Starting").Should().Succeed().
			ContainLines(1).
			ContainLine(0, "NAME")

		test.Machine("ls --filter state=Error").Should().Succeed().
			ContainLines(1).
			ContainLine(0, "NAME")
	})

	test.Run("ls: filter on name", func() {
		test.Machine("ls --filter name=testmachine2").Should().Succeed().
			ContainLines(2).
			ContainLine(0, "NAME").
			ContainLine(1, "testmachine2")

		test.Machine("ls -q --filter name=testmachine3").Should().Succeed().
			ContainLines(1).
			EqualLine(0, "testmachine3")
	})

	test.Run("ls: filter on name with regex", func() {
		test.Machine("ls --filter name=^t.*e[3-5]").Should().Succeed().
			ContainLines(4).
			ContainLine(0, "NAME").
			ContainLine(1, "testmachine3").
			ContainLine(2, "testmachine4").
			ContainLine(3, "testmachine5")

		test.Machine("ls -q --filter name=^t.*e[45]").Should().Succeed().
			ContainLines(2).
			EqualLine(0, "testmachine4").
			EqualLine(1, "testmachine5")
	})

	test.Run("setup swarm", func() {
		test.Machine("create -d none --url tcp://127.0.0.1:2375 --swarm --swarm-master --swarm-discovery token://deadbeef testswarm").Should().Succeed()
		test.Machine("create -d none --url tcp://127.0.0.1:2375 --swarm --swarm-discovery token://deadbeef testswarm2").Should().Succeed()
		test.Machine("create -d none --url tcp://127.0.0.1:2375 --swarm --swarm-discovery token://deadbeef testswarm3").Should().Succeed()
	})

	test.Run("ls: filter on swarm", func() {
		test.Machine("ls --filter swarm=testswarm").Should().Succeed().
			ContainLines(4).
			ContainLine(0, "NAME").
			ContainLine(1, "testswarm").
			ContainLine(2, "testswarm2").
			ContainLine(3, "testswarm3")
	})

	test.Run("ls: multi filter", func() {
		test.Machine("ls -q --filter swarm=testswarm --filter name=^t.*e --filter driver=none --filter state=Running").Should().Succeed().
			ContainLines(3).
			EqualLine(0, "testswarm").
			EqualLine(1, "testswarm2").
			EqualLine(2, "testswarm3")
	})

	test.Run("ls: format on driver", func() {
		test.Machine("ls --format {{.DriverName}}").Should().Succeed().
			ContainLines(8).
			EqualLine(0, "none").
			EqualLine(1, "none").
			EqualLine(2, "none").
			EqualLine(3, "none").
			EqualLine(4, "none").
			EqualLine(5, "none").
			EqualLine(6, "none").
			EqualLine(7, "none")
	})

	test.Run("ls: format on name and driver", func() {
		test.Machine("ls --format 'table {{.Name}}: {{.DriverName}}'").Should().Succeed().
			ContainLines(9).
			ContainLine(0, "NAME").
			EqualLine(1, "testmachine: none").
			EqualLine(2, "testmachine2: none").
			EqualLine(3, "testmachine3: none").
			EqualLine(4, "testmachine4: none").
			EqualLine(5, "testmachine5: none").
			EqualLine(6, "testswarm: none").
			EqualLine(7, "testswarm2: none").
			EqualLine(8, "testswarm3: none")
	})
}
