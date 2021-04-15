# What is these scripts
These scripts are for the benchmark of cpu usage, minikube vs kind vs k3d.   

 * `benchmark_local_k8s.sh`: take benchmark for cpu usage. This will take long to take place  
 * `update_summary.sh`: create one summary csv file of each drivers and products
 * `chart.go`: create bar chart graph as a png file

# How to use these scripts
 
```
cd <Top of minikube directory>
make cpu-benchmark
```

After running `make cpu-benchmark`, the png file of the bar chart graph will be generated.  
If you update the benchmark results to [our website](https://minikube.sigs.k8s.io/docs/benchmarks/), please commit this change.

```
git status
git add <Changed png file>
git commit
```