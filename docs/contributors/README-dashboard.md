# How to access Dashboard of Kubernete?

## First, see kube-system "pod"
```
sudo kubectl get pods --namespace=kube-system
```

You will see the list of kube-system pods
```
NAME                                    READY     STATUS    RESTARTS   AGE
etcd-minikube                           1/1       Running   2          1h
kube-addon-manager-minikube             1/1       Running   2          1h
kube-apiserver-minikube                 1/1       Running   2          59m
kube-controller-manager-minikube        1/1       Running   2          1h
kube-dns-86f4d74b45-nclqt               3/3       Running   4          1h
kube-proxy-jzzck                        1/1       Running   1          1h
kube-scheduler-minikube                 1/1       Running   2          1h
kubernetes-dashboard-5498ccf677-d9wg8   1/1       Running   1          1h
storage-provisioner                     1/1       Running   2          1h
```

## Second, see kube-system "svc"
```
sudo kubectl get svc --namespace=kube-system
```
You will see the list of svc
```
NAME                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)         AGE
kube-dns               ClusterIP   10.96.0.10     <none>        53/UDP,53/TCP   1h
kubernetes-dashboard   NodePort    10.104.9.212   <none>        80:30000/TCP    1h
```

## Third, find Dashboard UI web port
Find/grab the Dashboard Web UI port for Web browser to 
```
sudo kubectl get svc --namespace=kube-system|grep 'kubernetes-dashboard' | cut -d':' -f2 | cut -d'/' -f1
```
You will see the web port
```
30000
```

## Finally, launch your Firefox or Chrome
```
/usr/bin/google-chrome http://127.0.0.1:30000/
```

## Reference
* https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/#deploying-the-dashboard-ui


