# depends-on
A small container for Kubernetes that waits for the specified(as arguments) services have at least 
one ready(with readiness probe passed) pod.

Can be used as an init container or Helm's pre-install hook job
Since the service account with which this container is started must have "get" and "list" capabilities 
for "pods" and "services", the second approach is more preferable from security perspective. In case
of an init container, the main container also will be granted with these capabilities because 
"serviceAccountName" can be specified on a pod level only.

# Building
```
docker build . -t meshok0/depends-on:0.0.1
docker push  meshok0/depends-on:0.0.1 
```

# Installing to Kubernetes cluster
See example at example.yaml 
```
kubectl apply -f example.yaml --namespace=depon-test 
```

# Observing logs
```
kubectl logs second-7f86b9fff6-xjtjb -c second-depon  --namespace=depon-test -f
```

# Cleaning up
```
kubectl delete -f example.yaml --namespace=depon-test 
```