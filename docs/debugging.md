# Debugging Issues With Minikube

To debug issues with minikube (not *Kubernetes* but **minikube** itself), you can use the `-v` flag to see debug level info.  The specified values for `-v` will do the following (the values are all encompassing in that higher values will give you all lower value outputs as well):

* `--v=0` will output **INFO** level logs
* `--v=1` will output **WARNING** level logs
* `--v=2` will output **ERROR** level logs
* `--v=3` will output *libmachine* logging
* `--v=7` will output *libmachine --debug* level logging

Example:
`minikube start --v=1` Will start minikube and output all warnings to stdout.

If you need to access additional tools for debugging, minikube also includes the [CoreOS toolbox](https://github.com/coreos/toolbox)

You can ssh into the toolbox and access these additional commands using:
`minikube ssh toolbox`
