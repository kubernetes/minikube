To use baremetal driver (native driver) verify that your operating system is Linux and also have 'systemd' installed.

```shell
pidof systemd && echo "yes" || echo "no"
```
If the above command outputs "no":
Your system is not suitable for native driver.