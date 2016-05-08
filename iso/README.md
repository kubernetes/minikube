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
