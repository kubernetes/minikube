
# minikube Environment Variables

## Config option variables

minikube supports passing environment variables instead of flags for every value listed in `minikube config list`.  This is done by passing an environment variable with the prefix `MINIKUBE_`.

For example the `minikube start --iso-url="$ISO_URL"` flag can also be set by setting the `MINIKUBE_ISO_URL="$ISO_URL"` environment variable.

## Other variables

Some features can only be accessed by environment variables, here is a list of these features:

* **MINIKUBE_HOME** - (string) sets the path for the .minikube directory that minikube uses for state/configuration

* **MINIKUBE_IN_STYLE** - (bool) manually sets whether or not emoji and colors should appear in minikube. Set to false or 0 to disable this feature, true or 1 to force it to be turned on.

* **MINIKUBE_WANTUPDATENOTIFICATION** - (bool) sets whether the user wants an update notification for new minikube versions

* **MINIKUBE_REMINDERWAITPERIODINHOURS** - (int) sets the number of hours to check for an update notification

* **MINIKUBE_WANTKUBECTLDOWNLOADMSG** - (bool) sets whether minikube should tell a user that `kubectl` cannot be found on there path
* **MINIKUBE_WANTNONEDRIVERWARNING** - (bool) sets whether minikube should warn a user about running the 'none' driver

* **MINIKUBE_ENABLE_PROFILING** - (int, `1` enables it) enables trace profiling to be generated for minikube

## Making these values permanent

To make the exported variables permanent:

* Linux and macOS: Add these declarations to `~/.bashrc` or wherever your shells environment variables are stored.
* Windows: Add these declarations via [system settings](https://support.microsoft.com/en-au/help/310519/how-to-manage-environment-variables-in-windows-xp) or using [setx](https://stackoverflow.com/questions/5898131/set-a-persistent-environment-variable-from-cmd-exe)

### Example: Disabling emoji

```shell
export MINIKUBE_IN_STYLE=false
minikube start
```

### Example: Profiling

```shell
MINIKUBE_ENABLE_PROFILING=1 minikube start
```

Output:

``` text
2017/01/09 13:18:00 profile: cpu profiling enabled, /tmp/profile933201292/cpu.pprof
Starting local Kubernetes cluster...
Kubectl is now configured to use the cluster.
2017/01/09 13:19:06 profile: cpu profiling disabled, /tmp/profile933201292/cpu.pprof
```

Examine the cpu profiling results:

```shell
go tool pprof  /tmp/profile933201292/cpu.pprof
```
