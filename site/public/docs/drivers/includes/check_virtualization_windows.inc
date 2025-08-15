To check if virtualization is supported, run the following command on your Windows terminal or command prompt.

```shell
systeminfo
```
If you see the following output, virtualization is supported:

```shell
Hyper-V Requirements:     VM Monitor Mode Extensions: Yes
                          Virtualization Enabled In Firmware: Yes
                          Second Level Address Translation: Yes
                          Data Execution Prevention Available: Yes
```

If you see the following output, your system already has a Hypervisor installed and you can skip the next step.

```shell
Hyper-V Requirements:     A hypervisor has been detected.
```