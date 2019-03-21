# Accessing Host Resources From Inside A Pod

## When you have a VirtualBox driver

In order to access host resources from inside a pod, run the following command to determine the host IP you can use:

```shell
ip addr
```

The IP address under `vboxnet1` is the IP that you need to access the host from within a pod.
