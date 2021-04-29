---
title: "Integration Tests"
description: >
  All minikube integration tests
---


## TestDownloadOnly
TestDownloadOnly makes sure the --download-only parameter in minikube start caches the appropriate images and tarballs.

## TestDownloadOnlyKic
TestDownloadOnlyKic makes sure --download-only caches the docker driver images as well.

## TestOffline
TestOffline makes sure minikube works without internet, once the user has cached the necessary images.
This test has to run after TestDownloadOnly.

## TestAddons
TestAddons tests addons that require no special environment in parallel

#### validateIngressAddon
validateIngressAddon tests the ingress addon by deploying a default nginx pod

#### validateRegistryAddon
validateRegistryAddon tests the registry addon

#### validateMetricsServerAddon
validateMetricsServerAddon tests the metrics server addon by making sure "kubectl top pods" returns a sensible result

#### validateHelmTillerAddon
validateHelmTillerAddon tests the helm tiller addon by running "helm version" inside the cluster

#### validateOlmAddon
validateOlmAddon tests the OLM addon

#### validateCSIDriverAndSnapshots
validateCSIDriverAndSnapshots tests the csi hostpath driver by creating a persistent volume, snapshotting it and restoring it.

#### validateGCPAuthAddon
validateGCPAuthAddon tests the GCP Auth addon with either phony or real credentials and makes sure the files are mounted into pods correctly

## TestCertOptions
TestCertOptions makes sure minikube certs respect the --apiserver-ips and --apiserver-names parameters

## TestDockerFlags
TestDockerFlags makes sure the --docker-env and --docker-opt parameters are respected

## TestForceSystemdFlag
TestForceSystemdFlag tests the --force-systemd flag, as one would expect.

#### validateDockerSystemd
validateDockerSystemd makes sure the --force-systemd flag worked with the docker container runtime

#### validateContainerdSystemd
validateContainerdSystemd makes sure the --force-systemd flag worked with the containerd container runtime

## TestForceSystemdEnv
TestForceSystemdEnv makes sure the MINIKUBE_FORCE_SYSTEMD environment variable works just as well as the --force-systemd flag

## TestKVMDriverInstallOrUpdate
TestKVMDriverInstallOrUpdate makes sure our docker-machine-driver-kvm2 binary can be installed properly

## TestHyperKitDriverInstallOrUpdate
TestHyperKitDriverInstallOrUpdate makes sure our docker-machine-driver-hyperkit binary can be installed properly

## TestHyperkitDriverSkipUpgrade
TestHyperkitDriverSkipUpgrade makes sure our docker-machine-driver-hyperkit binary can be installed properly

## TestErrorSpam
TestErrorSpam asserts that there are no unexpected errors displayed in minikube command outputs.

## TestFunctional
TestFunctional are functionality tests which can safely share a profile in parallel

#### validateNodeLabels
validateNodeLabels checks if minikube cluster is created with correct kubernetes's node label

#### validateLoadImage
validateLoadImage makes sure that `minikube load image` works as expected

#### validateRemoveImage
validateRemoveImage makes sures that `minikube rm image` works as expected

#### validateBuildImage
validateBuildImage makes sures that `minikube image build` works as expected

#### validateDockerEnv
check functionality of minikube after evaling docker-env

#### validateStartWithProxy
validateStartWithProxy makes sure minikube start respects the HTTP_PROXY environment variable

#### validateAuditAfterStart
validateAuditAfterStart makes sure the audit log contains the correct logging after minikube start

#### validateSoftStart
validateSoftStart validates that after minikube already started, a "minikube start" should not change the configs.

#### validateKubeContext
validateKubeContext asserts that kubectl is properly configured (race-condition prone!)

#### validateKubectlGetPods
validateKubectlGetPods asserts that `kubectl get pod -A` returns non-zero content

#### validateMinikubeKubectl
validateMinikubeKubectl validates that the `minikube kubectl` command returns content

#### validateMinikubeKubectlDirectCall
validateMinikubeKubectlDirectCall validates that calling minikube's kubectl

#### validateExtraConfig
validateExtraConfig verifies minikube with --extra-config works as expected

#### validateComponentHealth
validateComponentHealth asserts that all Kubernetes components are healthy
NOTE: It expects all components to be Ready, so it makes sense to run it close after only those tests that include '--wait=all' start flag

#### validateStatusCmd
validateStatusCmd makes sure minikube status outputs correctly

#### validateDashboardCmd
validateDashboardCmd asserts that the dashboard command works

#### validateDryRun
validateDryRun asserts that the dry-run mode quickly exits with the right code

#### validateCacheCmd
validateCacheCmd tests functionality of cache command (cache add, delete, list)

#### validateConfigCmd
validateConfigCmd asserts basic "config" command functionality

#### validateLogsCmd
validateLogsCmd asserts basic "logs" command functionality

#### validateProfileCmd
validateProfileCmd asserts "profile" command functionality

#### validateServiceCmd
validateServiceCmd asserts basic "service" command functionality

#### validateAddonsCmd
validateAddonsCmd asserts basic "addon" command functionality

#### validateSSHCmd
validateSSHCmd asserts basic "ssh" command functionality

#### validateCpCmd
validateCpCmd asserts basic "cp" command functionality

#### validateMySQL
validateMySQL validates a minimalist MySQL deployment

#### validateFileSync
validateFileSync to check existence of the test file

#### validateCertSync
validateCertSync to check existence of the test certificate

#### validateUpdateContextCmd
validateUpdateContextCmd asserts basic "update-context" command functionality

#### validateMountCmd
validateMountCmd verifies the minikube mount command works properly

#### validatePersistentVolumeClaim
validatePersistentVolumeClaim makes sure PVCs work properly

#### validateTunnelCmd
validateTunnelCmd makes sure the minikube tunnel command works as expected

#### validateTunnelStart
validateTunnelStart starts `minikube tunnel`

#### validateServiceStable
validateServiceStable starts nginx pod, nginx service and waits nginx having loadbalancer ingress IP

#### validateAccessDirect
validateAccessDirect validates if the test service can be accessed with LoadBalancer IP from host

#### validateDNSDig
validateDNSDig validates if the DNS forwarding works by dig command DNS lookup
NOTE: DNS forwarding is experimental: https://minikube.sigs.k8s.io/docs/handbook/accessing/#dns-resolution-experimental

#### validateDNSDscacheutil
validateDNSDscacheutil validates if the DNS forwarding works by dscacheutil command DNS lookup
NOTE: DNS forwarding is experimental: https://minikube.sigs.k8s.io/docs/handbook/accessing/#dns-resolution-experimental

#### validateAccessDNS
validateAccessDNS validates if the test service can be accessed with DNS forwarding from host
NOTE: DNS forwarding is experimental: https://minikube.sigs.k8s.io/docs/handbook/accessing/#dns-resolution-experimental

#### validateTunnelDelete
validateTunnelDelete stops `minikube tunnel`

## TestGuestEnvironment
TestGuestEnvironment verifies files and packges installed inside minikube ISO/Base image

## TestGvisorAddon
TestGvisorAddon tests the functionality of the gVisor addon

## TestJSONOutput
TestJSONOutput makes sure json output works properly for the start, pause, unpause, and stop commands

#### validateDistinctCurrentSteps
 validateDistinctCurrentSteps makes sure each step has a distinct step number

#### validateIncreasingCurrentSteps
validateIncreasingCurrentSteps verifies that for a successful minikube start, 'current step' should be increasing

## TestErrorJSONOutput
TestErrorJSONOutput makes sure json output can print errors properly

## TestKicCustomNetwork
TestKicCustomNetwork verifies the docker driver works with a custom network

## TestKicExistingNetwork
TestKicExistingNetwork verifies the docker driver and run with an existing network

## TestingKicBaseImage
TestingKicBaseImage will return true if the integraiton test is running against a passed --base-image flag

## TestMultiNode
TestMultiNode tests all multi node cluster functionality

#### validateMultiNodeStart
validateMultiNodeStart makes sure a 2 node cluster can start

#### validateAddNodeToMultiNode
validateAddNodeToMultiNode uses the minikube node add command to add a node to an existing cluster

#### validateProfileListWithMultiNode
validateProfileListWithMultiNode make sure minikube profile list outputs correct with multinode clusters

#### validateStopRunningNode
validateStopRunningNode tests the minikube node stop command

#### validateStartNodeAfterStop
validateStartNodeAfterStop tests the minikube node start command on an existing stopped node

#### validateStopMultiNodeCluster
validateStopMultiNodeCluster runs minikube stop on a multinode cluster

#### validateRestartMultiNodeCluster
validateRestartMultiNodeCluster verifies a soft restart on a multinode cluster works

#### validateDeleteNodeFromMultiNode
validateDeleteNodeFromMultiNode tests the minikube node delete command

#### validateNameConflict
validateNameConflict tests that the node name verification works as expected

#### validateDeployAppToMultiNode
validateDeployAppToMultiNode deploys an app to a multinode cluster and makes sure all nodes can serve traffic

## TestNetworkPlugins
TestNetworkPlugins tests all supported CNI options
Options tested: kubenet, bridge, flannel, kindnet, calico, cilium
Flags tested: enable-default-cni (legacy), false (CNI off), auto-detection

## TestChangeNoneUser
TestChangeNoneUser tests to make sure the CHANGE_MINIKUBE_NONE_USER environemt variable is respected
and changes the minikube file permissions from root to the correct user.

## TestPause
TestPause tests minikube pause functionality

#### validateFreshStart
validateFreshStart just starts a new minikube cluster

#### validateStartNoReconfigure
validateStartNoReconfigure validates that starting a running cluster does not invoke reconfiguration

#### validatePause
validatePause runs minikube pause

#### validateUnpause
validateUnpause runs minikube unpause

#### validateDelete
validateDelete deletes the unpaused cluster

#### validateVerifyDeleted
validateVerifyDeleted makes sure no left over left after deleting a profile such as containers or volumes

#### validateStatus
validateStatus makes sure paused clusters show up in minikube status correctly

## TestDebPackageInstall
TestPackageInstall tests installation of .deb packages with minikube itself and with kvm2 driver
on various debian/ubuntu docker images

## TestPreload
TestPreload verifies the preload tarballs get pulled in properly by minikube

## TestScheduledStopWindows
TestScheduledStopWindows tests the schedule stop functionality on Windows

## TestScheduledStopUnix
TestScheduledStopWindows tests the schedule stop functionality on Unix

## TestSkaffold
TestSkaffold makes sure skaffold run can be run with minikube

## TestStartStop
TestStartStop tests starting, stopping and restarting a minikube clusters with various Kubernetes versions and configurations
The oldest supported, newest supported and default Kubernetes versions are always tested.

#### validateFirstStart
validateFirstStart runs the initial minikube start

#### validateDeploying
validateDeploying deploys an app the minikube cluster

#### validateStop
validateStop tests minikube stop

#### validateEnableAddonAfterStop
validateEnableAddonAfterStop makes sure addons can be enabled on a stopped cluster

#### validateSecondStart
validateSecondStart verifies that starting a stopped cluster works

#### validateAppExistsAfterStop
validateAppExistsAfterStop verifies that a user's app will not vanish after a minikube stop

#### validateAddonAfterStop
validateAddonAfterStop validates that an addon which was enabled when minikube is stopped will be enabled and working..

#### validateKubernetesImages
validateKubernetesImages verifies that a restarted cluster contains all the necessary images

#### validatePauseAfterStart
validatePauseAfterStart verifies that minikube pause works

## TestInsufficientStorage
TestInsufficientStorage makes sure minikube status displays the correct info if there is insufficient disk space on the machine

## TestRunningBinaryUpgrade
TestRunningBinaryUpgrade upgrades a running legacy cluster to minikube at HEAD

## TestStoppedBinaryUpgrade
TestStoppedBinaryUpgrade starts a legacy minikube, stops it, and then upgrades to minikube at HEAD

## TestKubernetesUpgrade
TestKubernetesUpgrade upgrades Kubernetes from oldest to newest

## TestMissingContainerUpgrade
TestMissingContainerUpgrade tests a Docker upgrade where the underlying container is missing

TEST COUNT: 111
