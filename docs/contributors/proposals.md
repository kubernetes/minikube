# Minikube Proposals
This document contains proposed functionality and rough design docs for said
functionality.

## OS X Menubar App 
The OS X Menubar App (proposed here: https://github.com/kubernetes/minikube/issues/1866) 
is designed to simplify basic minikube and k8s tasks for end users of minikube 
installations who are not comfortable with complex command line options. The 
app will sit in the menubar and allow a user easily accomplish the following 
tasks: 

* Start and Stop minikube (and show status of minikube/localkube)
* Launch the k8s dashboard
* See a list of pods and their status (red or green)
* SSH into any pod
* Tail logs in the terminal from any pod 
* Kill a pod 
* Configure minikube startup options
* Start minikube on boot
* Restart minikube if localkube dies (https://github.com/kubernetes/minikube/issues/1839)
* Gather bug report information and launch a browser to file an issue 
* Link to the minikube docs/project 
* Spin up a deployment/etc by loading a folder/yaml file


### App Design 
The app will be a native Swift application and will be added to a new folder 
inside `./installers/darwin/menubar_app/`
