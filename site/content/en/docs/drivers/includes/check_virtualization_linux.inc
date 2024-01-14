To use VM drivers, verify that your system has virtualization support enabled:

```shell
grep -E -q 'vmx|svm' /proc/cpuinfo && echo yes || echo no
```

If the above command outputs "no":

- If you are running within a VM, your hypervisor does not allow nested virtualization. You will need to use the *None (bare-metal)* driver
- If you are running on a physical machine, ensure that your BIOS has hardware virtualization enabled

