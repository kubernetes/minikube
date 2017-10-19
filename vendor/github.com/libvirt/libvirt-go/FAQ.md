#libvirt-go

##FAQ - Frequently asked questions

If your question is a good one, please ask it as a well-formatted patch to this
repository, and we'll merge it along with the answer.

###Why does this fail when added to my project in travis?

This lib requires a newish version of the libvirt-dev library to compile. These
are only available in the newer travis environment. You can add:

```
sudo: true
dist: trusty
install: sudo apt-get install -y libvirt-dev
```

to your `.travis.yaml` file to avoid these errors.
