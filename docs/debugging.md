### Debugging Issues With Minikube
To debug issues with minikube (not Kubernetes but minikube itself), you can use the -v flag to see debug level info.  The specified values for v will do the following (the values are all encompassing in that higher values will give you all lower value outputs as well):
* --v=0 INFO level logs
* --v=1 WARNING level logs
* --v=2 ERROR level logs
* --v=3 libmachine logging
* --v=7 libmachine --debug level logging

If you need to access additional tools for debugging, minikube also includes the [CoreOS toolbox](https://github.com/coreos/toolbox)


You can ssh into the toolbox and access these additional commands using:
`minikube ssh toolbox`
