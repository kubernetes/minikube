package tester

import (
	"os"

	"k8s.io/klog/v2"
)

var _ MiniTestTester = &KVMIntegrationTester{}

type KVMIntegrationTester struct {
}

// Run implements MiniTestTester.
func (k *KVMIntegrationTester) Run(runner MiniTestRunner) error {

	if up, err := runner.IsUp(); err != nil || !up {
		klog.Errorf("tester: deployed environment is not up: %v", err)
	}

	if err := runner.SyncToRemote(".", "~/minikube"); err != nil {
		klog.Errorf("failed to sync file in docker deployer: %v", err)
	}
	// install docker and libvirtd first then run the test in a new shell
	if err := runner.Execute("cd minikube && ./hack/prow/linux_integration_kvm_pre.sh"); err != nil {
		klog.Errorf("failed to install docker in env: %v", err)
		return err
	}
	if err := runner.Execute("cd minikube && ./hack/prow/linux_integration_kvm.sh"); err != nil {
		klog.Errorf("failed to execute command in env: %v", err)
		// don't return here, we still want to collect the test reports
	}
	artifactLocation := os.Getenv("ARTIFACTS")
	klog.Infof("copying to %s", artifactLocation)

	if err := runner.SyncToHost("~/minikube/testout.txt", artifactLocation); err != nil {
		klog.Errorf("failed to sync testout.txt from deployer: %v", err)
		return err
	}
	if err := runner.SyncToHost("~/minikube/test.json", artifactLocation); err != nil {
		klog.Errorf("failed to sync test.json in from deployer: %v", err)
		return err
	}

	if err := runner.SyncToHost("~/minikube/junit-unit.xml", artifactLocation); err != nil {
		klog.Errorf("failed to sync junit-unit.xml in from deployer: %v", err)
		return err
	}

	return nil

}
