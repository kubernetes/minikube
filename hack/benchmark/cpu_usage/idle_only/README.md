# What is these scripts
These scripts are for the benchmark of cpu usage, minikube vs kind vs k3d.   

 * `benchmark_local_k8s.sh`: take benchmark for cpu usage. This will take long to take place  
 * `update_summary.sh`: create one summary csv file of each drivers and products
 * `chart.go`: create bar chart graph as a png file
 
In `benchmark_local_k8s.sh`, we compare minikube drivers(docker, hyperkit, virtualbox) and kind, k3d, Docker for Mac Kubernetes in case of macOS.   
In `benchmark_local_k8s.sh`, we compare minikube drivers(docker, kvm2, virtualbox) and kind, k3d in case of Linux.   
`benchmark_local_k8s.sh` take these steps to measure idle usage after start-up.   

  1. start each local kubernetes tool
  2. measure its cpu usage with [cstat](https://github.com/tstromberg/cstat)

# How to use these scripts
 
```
cd <Top of minikube directory>
make cpu-benchmark-idle
```

After running `make cpu-benchmark-idle`, the png file of the bar chart graph will be generated.  
If you update the benchmark results to [our website](https://minikube.sigs.k8s.io/docs/benchmarks/), please commit this change.

```
git status
git add <Changed png file>
git commit
```