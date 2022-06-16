# What is these scripts
These scripts are for the benchmark of cpu usage, minikube vs kind vs k3d.   

 * `benchmark_local_k8s.sh`: take benchmark for cpu usage. This will take long to take place  
 * `update_summary.sh`: create one summary csv file of each drivers and products
 * `chart.go`: create bar chart graph as a png file
 
In `benchmark_local_k8s.sh`, we compare minikube drivers(hyperkit, virtualbox, docker with auto-pause addon) and kind, k3d, Docker for Mac Kubernetes in case of macOS.   
In `benchmark_local_k8s.sh`, we compare minikube drivers(kvm2, virtualbox, docker with auto-pause addon) and kind, k3d in case of Linux.   
`benchmark_local_k8s.sh` take these steps to measure `auto-pause` vs. `non auto-pause`.   

 1. start each local kubernetes tool
 2. deploy sample application(nginx deployment) to each tool
 3. wait 1 minute without anything
 4. measure No.3 idle CPU usage with [cstat](https://github.com/tstromberg/cstat)
 5. if tool is minikube, enable auto-pause addon which pause control plane
 6. if tool is minikube, wait 1 minute so that control plane will become Paused status(It takes 1 minute to become Pause status from Stopped status)  
 7. if tool is minikube, verify if minikube control plane is paused
 8. if tool is minikube, wait 3 minute without anything
 9. if tool is minikube, measure No.8 idle CPU usage with [cstat](https://github.com/tstromberg/cstat)

No.1-4: Initial start CPU usage with sample(nginx) deployment   
No.5-9: Auto Paused CPU usage with sample(nginx) deployment    

# How to use these scripts
 
```
cd <Top of minikube directory>
make cpu-benchmark-benchmark-autopause
```

After running `make cpu-benchmark-autopause`, the png file of the bar chart graph will be generated.  
If you update the benchmark results to [our website](https://minikube.sigs.k8s.io/docs/benchmarks/), please commit this change.

```
git status
git add <Changed png file>
git commit
```
