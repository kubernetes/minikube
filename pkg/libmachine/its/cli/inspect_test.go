package cli

import (
	"testing"

	"k8s.io/minikube/pkg/libmachine/its"
)

func TestInspect(t *testing.T) {
	test := its.NewTest(t)
	defer test.TearDown()

	test.Run("inspect: show error in case of no args", func() {
		test.Machine("inspect").Should().Fail(`Error: No machine name(s) specified and no "default" machine exists`)
	})
}
