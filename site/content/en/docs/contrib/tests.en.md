---
title: "List of Integration Test Cases"
description: >
  Auto generated list of all minikube integration tests and what they do.
---


## TestDownloadOnly
makes sure the --download-only parameter in minikube start caches the appropriate images and tarballs.

## TestDownloadOnlyKic
makes sure --download-only caches the docker driver images as well.

## TestBinaryMirror
tests functionality of --binary-mirror flag

## TestOffline
makes sure minikube works without internet, once the user has cached the necessary images.
This test has to run after TestDownloadOnly.

## TestAddons
tests addons that require no special environment in parallel

#### validateIngressAddon
tests the ingress addon by deploying a default nginx pod

#### validateRegistryAddon
tests the registry addon

#### validateMetricsServerAddon
tests the metrics server addon by making sure "kubectl top pods" returns a sensible result

#### validateHelmTillerAddon
tests the helm tiller addon by running "helm version" inside the cluster

#### validateOlmAddon
tests the OLM addon

#### validateCSIDriverAndSnapshots
tests the csi hostpath driver by creating a persistent volume, snapshotting it and restoring it.

#### validateGCPAuthNamespaces
validates that newly created namespaces contain the gcp-auth secret.

#### validateGCPAuthAddon
tests the GCP Auth addon with either phony or real credentials and makes sure the files are mounted into pods correctly

#### validateHeadlampAddon

#### validateInspektorGadgetAddon
tests the inspektor-gadget addon by ensuring the pod has come up and addon disables

#### validateCloudSpannerAddon
tests the cloud-spanner addon by ensuring the deployment and pod come up and addon disables

#### validateLocalPathAddon
tests the functionality of the storage-provisioner-rancher addon

#### validateEnablingAddonOnNonExistingCluster
tests enabling an addon on a non-existing cluster

#### validateDisablingAddonOnNonExistingCluster
tests disabling an addon on a non-existing cluster

#### validateNvidiaDevicePlugin
tests the nvidia-device-plugin addon by ensuring the pod comes up and the addon disables

#### validateYakdAddon

## TestCertOptions
makes sure minikube certs respect the --apiserver-ips and --apiserver-names parameters

## TestCertExpiration
makes sure minikube can start after its profile certs have expired.
It does this by configuring minikube certs to expire after 3 minutes, then waiting 3 minutes, then starting again.
It also makes sure minikube prints a cert expiration warning to the user.

## TestDockerFlags
makes sure the --docker-env and --docker-opt parameters are respected

## TestForceSystemdFlag
tests the --force-systemd flag, as one would expect.

#### validateDockerSystemd
makes sure the --force-systemd flag worked with the docker container runtime

#### validateContainerdSystemd
makes sure the --force-systemd flag worked with the containerd container runtime

#### validateCrioSystemd
makes sure the --force-systemd flag worked with the cri-o container runtime

## TestForceSystemdEnv
makes sure the MINIKUBE_FORCE_SYSTEMD environment variable works just as well as the --force-systemd flag

## TestDockerEnvContainerd
makes sure that minikube docker-env command works when the runtime is containerd

## TestKVMDriverInstallOrUpdate
makes sure our docker-machine-driver-kvm2 binary can be installed properly

## TestHyperKitDriverInstallOrUpdate
makes sure our docker-machine-driver-hyperkit binary can be installed properly

## TestHyperkitDriverSkipUpgrade
makes sure our docker-machine-driver-hyperkit binary can be installed properly

## TestErrorSpam
asserts that there are no unexpected errors displayed in minikube command outputs.

## TestFunctional
are functionality tests which can safely share a profile in parallel

#### validateNodeLabels
checks if minikube cluster is created with correct kubernetes's node label

Steps:
- Get the node labels from the cluster with `kubectl get nodes`
- check if the node labels matches with the expected Minikube labels: `minikube.k8s.io/*`

#### validateImageCommands
runs tests on all the `minikube image` commands, ex. `minikube image load`, `minikube image list`, etc.

Steps:
- Make sure image building works by `minikube image build`
- Make sure image loading from Docker daemon works by `minikube image load --daemon`
- Try to load image already loaded and make sure `minikube image load --daemon` works
- Make sure a new updated tag works by `minikube image load --daemon`
- Make sure image saving works by `minikube image load --daemon`
- Make sure image removal works by `minikube image rm`
- Make sure image loading from file works by `minikube image load`
- Make sure image saving to Docker daemon works by `minikube image load`

Skips:
- Skips on `none` driver as image loading is not supported
- Skips on GitHub Actions and macOS as this test case requires a running docker daemon

#### validateDockerEnv
check functionality of minikube after evaluating docker-env

Steps:
- Run `eval $(minikube docker-env)` to configure current environment to use minikube's Docker daemon
- Run `minikube status` to get the minikube status
- Make sure minikube components have status `Running`
- Make sure `docker-env` has status `in-use`
- Run eval `$(minikube -p profile docker-env)` and check if we are point to docker inside minikube
- Make sure `docker images` hits the minikube's Docker daemon by check if `gcr.io/k8s-minikube/storage-provisioner` is in the output of `docker images`

Skips:
- Skips on `none` drive since `docker-env` is not supported
- Skips on non-docker container runtime

#### validatePodmanEnv
check functionality of minikube after evaluating podman-env

Steps:
- Run `eval $(minikube podman-env)` to configure current environment to use minikube's Podman daemon, and `minikube status` to get the minikube status
- Make sure minikube components have status `Running`
- Make sure `podman-env` has status `in-use`
- Run `eval $(minikube docker-env)` again and `docker images` to list the docker images using the minikube's Docker daemon
- Make sure `docker images` hits the minikube's Podman daemon by check if `gcr.io/k8s-minikube/storage-provisioner` is in the output of `docker images`

Skips:
- Skips on `none` drive since `podman-env` is not supported
- Skips on non-docker container runtime
- Skips on non-Linux platforms

#### validateStartWithProxy
makes sure minikube start respects the HTTP_PROXY environment variable

Steps:
- Start a local HTTP proxy
- Start minikube with the environment variable `HTTP_PROXY` set to the local HTTP proxy

#### validateStartWithCustomCerts
makes sure minikube start respects the HTTPS_PROXY environment variable and works with custom certs
a proxy is started by calling the mitmdump binary in the background, then installing the certs generated by the binary
mitmproxy/dump creates the proxy at localhost at port 8080
only runs on GitHub Actions for amd64 linux, otherwise validateStartWithProxy runs instead

#### validateAuditAfterStart
makes sure the audit log contains the correct logging after minikube start

Steps:
- Read the audit log file and make sure it contains the current minikube profile name

#### validateSoftStart
validates that after minikube already started, a `minikube start` should not change the configs.

Steps:
- The test `validateStartWithProxy` should have start minikube, make sure the configured node port is `8441`
- Run `minikube start` again as a soft start
- Make sure the configured node port is not changed

#### validateKubeContext
asserts that kubectl is properly configured (race-condition prone!)

Steps:
- Run `kubectl config current-context`
- Make sure the current minikube profile name is in the output of the command

#### validateKubectlGetPods
asserts that `kubectl get pod -A` returns non-zero content

Steps:
- Run `kubectl get po -A` to get all pods in the current minikube profile
- Make sure the output is not empty and contains `kube-system` components

#### validateMinikubeKubectl
validates that the `minikube kubectl` command returns content

Steps:
- Run `minikube kubectl -- get pods` to get the pods in the current minikube profile
- Make sure the command doesn't raise any error

#### validateMinikubeKubectlDirectCall
validates that calling minikube's kubectl

Steps:
- Run `kubectl get pods` by calling the minikube's `kubectl` binary file directly
- Make sure the command doesn't raise any error

#### validateExtraConfig
verifies minikube with --extra-config works as expected

Steps:
- The tests before this already created a profile
- Soft-start minikube with different `--extra-config` command line option
- Load the profile's config
- Make sure the specified `--extra-config` is correctly returned

#### validateComponentHealth
asserts that all Kubernetes components are healthy
NOTE: It expects all components to be Ready, so it makes sense to run it close after only those tests that include '--wait=all' start flag

Steps:
- Run `kubectl get po po -l tier=control-plane -n kube-system -o=json` to get all the Kubernetes conponents
- For each component, make sure the pod status is `Running`

#### validateStatusCmd
makes sure `minikube status` outputs correctly

Steps:
- Run `minikube status` with custom format `host:{{.Host}},kublet:{{.Kubelet}},apiserver:{{.APIServer}},kubeconfig:{{.Kubeconfig}}`
- Make sure `host`, `kublete`, `apiserver` and `kubeconfig` statuses are shown in the output
- Run `minikube status` again as JSON output
- Make sure `host`, `kublete`, `apiserver` and `kubeconfig` statuses are set in the JSON output

#### validateDashboardCmd
asserts that the dashboard command works

Steps:
- Run `minikube dashboard --url` to start minikube dashboard and return the URL of it
- Send a GET request to the dashboard URL
- Make sure HTTP status OK is returned

#### validateDryRun
asserts that the dry-run mode quickly exits with the right code

Steps:
- Run `minikube start --dry-run --memory 250MB`
- Since the 250MB memory is less than the required 2GB, minikube should exit with an exit code `ExInsufficientMemory`
- Run `minikube start --dry-run`
- Make sure the command doesn't raise any error

#### validateInternationalLanguage
asserts that the language used can be changed with environment variables

Steps:
- Set environment variable `LC_ALL=fr` to enable minikube translation to French
- Start minikube with memory of 250MB which is too little: `minikube start --dry-run --memory 250MB`
- Make sure the dry-run output message is in French

#### validateCacheCmd
tests functionality of cache command (cache add, delete, list)

Steps:
- Run `minikube cache add` and make sure we can add a remote image to the cache
- Run `minikube cache add` and make sure we can build and add a local image to the cache
- Run `minikube cache delete` and make sure we can delete an image from the cache
- Run `minikube cache list` and make sure we can list the images in the cache
- Run `minikube ssh sudo crictl images` and make sure we can list the images in the cache with `crictl`
- Delete an image from minikube node and run `minikube cache reload` to make sure the image is brought back correctly

#### validateConfigCmd
asserts basic "config" command functionality

Steps:
- Run `minikube config set/get/unset` to make sure configuration is modified correctly

#### validateLogsCmd
asserts basic "logs" command functionality

Steps:
- Run `minikube logs` and make sure the logs contains some keywords like `apiserver`, `Audit` and `Last Start`

#### validateLogsFileCmd
asserts "logs --file" command functionality

Steps:
- Run `minikube logs --file logs.txt` to save the logs to a local file
- Make sure the logs are correctly written

#### validateProfileCmd
asserts "profile" command functionality

Steps:
- Run `minikube profile lis` and make sure the command doesn't fail for the non-existent profile `lis`
- Run `minikube profile list --output json` to make sure the previous command doesn't create a new profile
- Run `minikube profile list` and make sure the profiles are correctly listed
- Run `minikube profile list -o JSON` and make sure the profiles are correctly listed as JSON output

#### validateServiceCmd
asserts basic "service" command functionality

#### validateServiceCmdDeployApp
Create a new `registry.k8s.io/echoserver` deployment

#### validateServiceCmdList
Run `minikube service list` to make sure the newly created service is correctly listed in the output

#### validateServiceCmdJSON
Run `minikube service list -o JSON` and make sure the services are correctly listed as JSON output

#### validateServiceCmdHTTPS
Run `minikube service` with `--https --url` to make sure the HTTPS endpoint URL of the service is printed

#### validateServiceCmdFormat
Run `minikube service` with `--url --format={{.IP}}` to make sure the IP address of the service is printed

#### validateServiceCmdURL
Run `minikube service` with a regular `--url` to make sure the HTTP endpoint URL of the service is printed

#### validateServiceCmdConnect

Steps:
- Create a new `registry.k8s.io/echoserver` deployment
- Run `minikube service` with a regular `--url` to make sure the HTTP endpoint URL of the service is printed
- Make sure we can hit the endpoint URL with an HTTP GET request

#### validateAddonsCmd
asserts basic "addon" command functionality

Steps:
- Run `minikube addons list` to list the addons in a tabular format
- Make sure `dashboard`, `ingress` and `ingress-dns` is listed as available addons
- Run `minikube addons list -o JSON` lists the addons in JSON format

#### validateSSHCmd
asserts basic "ssh" command functionality

Steps:
- Run `minikube ssh echo hello` to make sure we can SSH into the minikube container and run an command
- Run `minikube ssh cat /etc/hostname` as well to make sure the command is run inside minikube

#### validateCpCmd
asserts basic "cp" command functionality

Steps:
- Run `minikube cp ...` to copy a file to the minikube node
- Run `minikube ssh sudo cat ...` to print out the copied file within minikube
- make sure the file is correctly copied

Skips:
- Skips `none` driver since `cp` is not supported

#### validateMySQL
validates a minimalist MySQL deployment

Steps:
- Run `kubectl replace --force -f testdata/mysql/yaml`
- Wait for the `mysql` pod to be running
- Run `mysql -e show databases;` inside the MySQL pod to verify MySQL is up and running
- Retry with exponential backoff if failed, as `mysqld` first comes up without users configured. Scan for names in case of a reschedule.

Skips:
- Skips for ARM64 architecture since it's not supported by MySQL

#### validateFileSync
to check existence of the test file

Steps:
- Test files have been synced into minikube in the previous step `setupFileSync`
- Check the existence of the test file
- Make sure the file is correctly synced

Skips:
- Skips on `none` driver since SSH is not supported

#### validateCertSync
checks to make sure a custom cert has been copied into the minikube guest and installed correctly

Steps:
- Check both the installed & reference certs and make sure they are symlinked

#### validateNotActiveRuntimeDisabled
asserts that for a given runtime, the other runtimes are disabled, for example for `containerd` runtime, `docker` and `crio` needs to be not running

Steps:
- For each container runtime, run `minikube ssh sudo systemctl is-active ...` and make sure the other container runtimes are not running

#### validateUpdateContextCmd
asserts basic "update-context" command functionality

Steps:
- Run `minikube update-context`
- Make sure the context has been correctly updated by checking the command output

#### validateVersionCmd
asserts `minikube version` command works fine for both --short and --components

Steps:
- Run `minikube version --short` and make sure the returned version is a valid semver
- Run `minikube version --components` and make sure the component versions are returned

#### validateLicenseCmd
asserts that the `minikube license` command downloads and untars the licenses
Note: This test will fail on release PRs as the licenses file for the new version won't be uploaded at that point

#### validateInvalidService
makes sure minikube will not start a tunnel for an unavailable service that has no running pods

#### validateMountCmd
verifies the minikube mount command works properly
for the platforms that support it, we're testing:
- a generic 9p mount
- a 9p mount on a specific port
- cleaning-mechanism for profile-specific mounts

#### validatePersistentVolumeClaim
makes sure PVCs work properly

#### validateTunnelCmd
makes sure the minikube tunnel command works as expected

#### validateTunnelStart
starts `minikube tunnel`

#### validateNoSecondTunnel
ensures only 1 tunnel can run simultaneously

#### validateServiceStable
starts nginx pod, nginx service and waits nginx having loadbalancer ingress IP

#### validateAccessDirect
validates if the test service can be accessed with LoadBalancer IP from host

#### validateDNSDig
validates if the DNS forwarding works by dig command DNS lookup
NOTE: DNS forwarding is experimental: https://minikube.sigs.k8s.io/docs/handbook/accessing/#dns-resolution-experimental

#### validateDNSDscacheutil
validates if the DNS forwarding works by dscacheutil command DNS lookup
NOTE: DNS forwarding is experimental: https://minikube.sigs.k8s.io/docs/handbook/accessing/#dns-resolution-experimental

#### validateAccessDNS
validates if the test service can be accessed with DNS forwarding from host
NOTE: DNS forwarding is experimental: https://minikube.sigs.k8s.io/docs/handbook/accessing/#dns-resolution-experimental

#### validateTunnelDelete
stops `minikube tunnel`

## TestGuestEnvironment
verifies files and packages installed inside minikube ISO/Base image

## TestGvisorAddon
tests the functionality of the gVisor addon

## TestImageBuild
makes sure the 'minikube image build' command works fine

#### validateSetupImageBuild
starts a cluster for the image builds

#### validateNormalImageBuild
is normal test case for minikube image build, with -t parameter

#### validateNormalImageBuildWithSpecifiedDockerfile
is normal test case for minikube image build, with -t and -f parameter

#### validateImageBuildWithBuildArg
is a test case building with --build-opt

#### validateImageBuildWithBuildEnv
is a test case building with --build-env

#### validateImageBuildWithDockerIgnore
is a test case building with .dockerignore

## TestJSONOutput
makes sure json output works properly for the start, pause, unpause, and stop commands

#### validateDistinctCurrentSteps
makes sure each step has a distinct step number

#### validateIncreasingCurrentSteps
verifies that for a successful minikube start, 'current step' should be increasing

## TestErrorJSONOutput
makes sure json output can print errors properly

## TestKicCustomNetwork
verifies the docker driver works with a custom network

## TestKicExistingNetwork
verifies the docker driver and run with an existing network

## TestKicCustomSubnet
verifies the docker/podman driver works with a custom subnet

## TestKicStaticIP
starts minikube with the static IP flag

## TestingKicBaseImage
will return true if the integraiton test is running against a passed --base-image flag

## TestMinikubeProfile

## TestMountStart
tests using the mount command on start

#### validateStartWithMount
starts a cluster with mount enabled

#### validateMount
checks if the cluster has a folder mounted

#### validateMountStop
stops a cluster

#### validateRestart
restarts a cluster

## TestMultiNode
tests all multi node cluster functionality

#### validateMultiNodeStart
makes sure a 2 node cluster can start

#### validateAddNodeToMultiNode
uses the minikube node add command to add a node to an existing cluster

#### validateProfileListWithMultiNode
make sure minikube profile list outputs correct with multinode clusters

#### validateCopyFileWithMultiNode
validateProfileListWithMultiNode make sure minikube profile list outputs correct with multinode clusters

#### validateMultiNodeLabels
check if all node labels were configured correctly

Steps:
- Get the node labels from the cluster with `kubectl get nodes`
- check if all node labels matches with the expected Minikube labels: `minikube.k8s.io/*`

#### validateStopRunningNode
tests the minikube node stop command

#### validateStartNodeAfterStop
tests the minikube node start command on an existing stopped node

#### validateRestartKeepsNodes
restarts minikube cluster and checks if the reported node list is unchanged

#### validateStopMultiNodeCluster
runs minikube stop on a multinode cluster

#### validateRestartMultiNodeCluster
verifies a soft restart on a multinode cluster works

#### validateDeleteNodeFromMultiNode
tests the minikube node delete command

#### validateNameConflict
tests that the node name verification works as expected

#### validateDeployAppToMultiNode
deploys an app to a multinode cluster and makes sure all nodes can serve traffic

#### validatePodsPingHost
uses app previously deplyed by validateDeployAppToMultiNode to verify its pods, located on different nodes, can resolve "host.minikube.internal".

## TestNetworkPlugins
tests all supported CNI options
Options tested: kubenet, bridge, flannel, kindnet, calico, cilium
Flags tested: enable-default-cni (legacy), false (CNI off), auto-detection

#### validateFalseCNI
checks that minikube returns and error
if container runtime is "containerd" or "crio"
and --cni=false

#### validateHairpinMode
makes sure the hairpinning (https://en.wikipedia.org/wiki/Hairpinning) is correctly configured for given CNI
try to access deployment/netcat pod using external, obtained from 'netcat' service dns resolution, IP address
should fail if hairpinMode is off

## TestNoKubernetes
tests starting minikube without Kubernetes,
for use cases where user only needs to use the container runtime (docker, containerd, crio) inside minikube

#### validateStartNoK8sWithVersion
expect an error when starting a minikube cluster without kubernetes and with a kubernetes version.

Steps:
- start minikube with no kubernetes.

#### validateStartWithK8S
starts a minikube cluster with Kubernetes started/configured.

Steps:
- start minikube with Kubernetes.
- return an error if Kubernetes is not running.

#### validateStartWithStopK8s
starts a minikube cluster while stopping Kubernetes.

Steps:
- start minikube with no Kubernetes.
- return an error if Kubernetes is not stopped.
- delete minikube profile.

#### validateStartNoK8S
starts a minikube cluster without kubernetes started/configured

Steps:
- start minikube with no Kubernetes.

#### validateK8SNotRunning
validates that there is no kubernetes running inside minikube

#### validateStopNoK8S
validates that minikube is stopped after a --no-kubernetes start

#### validateProfileListNoK8S
validates that profile list works with --no-kubernetes

#### validateStartNoArgs
validates that minikube start with no args works.

## TestChangeNoneUser
tests to make sure the CHANGE_MINIKUBE_NONE_USER environment variable is respected
and changes the minikube file permissions from root to the correct user.

## TestPause
tests minikube pause functionality

#### validateFreshStart
just starts a new minikube cluster

#### validateStartNoReconfigure
validates that starting a running cluster does not invoke reconfiguration

#### validatePause
runs minikube pause

#### validateUnpause
runs minikube unpause

#### validateDelete
deletes the unpaused cluster

#### validateVerifyDeleted
makes sure no left over left after deleting a profile such as containers or volumes

#### validateStatus
makes sure paused clusters show up in minikube status correctly

## TestPreload
verifies the preload tarballs get pulled in properly by minikube

## TestScheduledStopWindows
tests the schedule stop functionality on Windows

## TestScheduledStopUnix
tests the schedule stop functionality on Unix

## TestSkaffold
makes sure skaffold run can be run with minikube

## TestStartStop
tests starting, stopping and restarting a minikube clusters with various Kubernetes versions and configurations
The oldest supported, newest supported and default Kubernetes versions are always tested.

#### validateFirstStart
runs the initial minikube start

#### validateDeploying
deploys an app the minikube cluster

#### validateEnableAddonWhileActive
makes sure addons can be enabled while cluster is active.

#### validateStop
tests minikube stop

#### validateEnableAddonAfterStop
makes sure addons can be enabled on a stopped cluster

#### validateSecondStart
verifies that starting a stopped cluster works

#### validateAppExistsAfterStop
verifies that a user's app will not vanish after a minikube stop

#### validateAddonAfterStop
validates that an addon which was enabled when minikube is stopped will be enabled and working..

#### validateKubernetesImages
verifies that a restarted cluster contains all the necessary images

#### validatePauseAfterStart
verifies that minikube pause works

## TestInsufficientStorage
makes sure minikube status displays the correct info if there is insufficient disk space on the machine

## TestRunningBinaryUpgrade
upgrades a running legacy cluster to minikube at HEAD

## TestStoppedBinaryUpgrade
starts a legacy minikube, stops it, and then upgrades to minikube at HEAD

## TestKubernetesUpgrade
upgrades Kubernetes from oldest to newest

## TestMissingContainerUpgrade
tests a Docker upgrade where the underlying container is missing

