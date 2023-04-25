package cli

import (
	"testing"

	"k8s.io/minikube/pkg/libmachine/its"
)

func TestStatus(t *testing.T) {
	test := its.NewTest(t)
	defer test.TearDown()

	test.Run("status: show error in case of no args", func() {
		test.Machine("status").Should().Fail(`Error: No machine name(s) specified and no "default" machine exists`)
	})
}
