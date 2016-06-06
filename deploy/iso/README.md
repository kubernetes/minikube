# Scripts for building a VM iso based on boot2docker

## Build instructions
./build.sh

## Test instructions
To manually try out the built iso, you can run the following commands to create a VM:

```shell
VBoxManage createvm --name testminikube --ostype "Linux_64" --register
VBoxManage storagectl testminikube --name "IDE Controller" --add ide
VBoxManage storageattach testminikube --storagectl "IDE Controller" --port 0 --device 0 --type dvddrive --medium ./minikube.iso
VBoxManage modifyvm testminikube --memory 1024 --vrde on --vrdeaddress 127.0.0.1 --vrdeport 3390 --vrdeauthtype null
```

Then use the VirtualBox gui to start and open a session.

## Release instructions
 * Build an iso following the above build instructions.
 * Test the iso with --iso-url=file:///$PATHTOISO.
 * Push the new iso to GCS, with a new name (minikube-0x.iso) with a command like this: `gsutil cp $PATHTOISO gs://$BUCKET`
 * Update the default URL in start.go.
