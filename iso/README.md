# Scripts for building a VM iso based on boot2docker

## Build instructions
./build.sh

## Test instructions
To manually try out the built iso, you can run the following commands to create a VM:

```shell
VBoxManage createvm --name testminikube --ostype "Linux_64" --register
VBoxManage storagectl foo --name "IDE Controller" --add ide
VBoxManage storageattach foo --storagectl "IDE Controller" --port 0 --device 0 --type dvddrive --medium ./minikube.iso
VBoxManage modifyvm foo --memory 1024 --vrde on --vrdeaddress 127.0.0.1 --vrdeport 3390 --vrdeauthtype null
```

Then use the VirtualBox gui to start and open a session.

## Release Instructions
minikube CLIs contain a pinned ISO version constant in constants.ISOVersion. The versionmap.json file is downloaded from
GCS and used to lookup the ISO URl for a given ISOVersion. This URL can be overriden with a flag for testing.

To release a new ISO:
 * Build it locally and test it with --iso-url=file:///$PATH_TO_ISO.
 * Upload it to GCS with a new name.
 * Add an entry to versionmap.json, commit that and send a PR.
 * Upload versionmap.json to GCS after the PR is merged.
