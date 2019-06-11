# The problem

Kubernetes and Helm don't have a built-in mechanism to deploy pods in a specific order.

If you've been working with dynamic/disposable/preview environments on Kubernetes you probably saw the following happen:

A new environment is created with Helm.

A dozen of new pods starts simultaneously (or in some random order because your cluster scales up).

Some pods, like MongoDB for example, should run and be available before all the other pods can start. Since it's not available just yet, it causes the other pods to crash and change status to `crashLoopBackOff`.

Now because of that behavior, the whole process of creating a new dev environment from scratch takes much longer than it should.

This small utility solves that problem and helps you minimize the time it takes to spin a new environment to the bare minimum!

# The solution

What if you could specify which services a pod should wait for before starting up? Or in short - the services it depends(-)on?

This small container allows you to add an `InitContainer` hook to all of your deployment YAML files to create a dependency relationship to other services in the same namespace:

```
initContainers:
- name: wait-for-services
  image: opsfleet/depends-on
  args:
  - "-service=mongodb"
  - "-service=rabbitmq"
  - "-service=elasticsearch"
```

## How it works

The logic isn't that complicated - the utility will block and periodically query the Kuberentes API until it finds at least one available endpoint for all the specified services.

If you've configured your readiness probes correctly, depends-on will wait until at least one pod for each service you've specified becomes available.

## Note: RBAC & Service account permissions

This utility was designed to be used in development and pre-production environments. You know, the envs where you can ease up the security a bit.

Since depend-on queries the Kubernets API, you'll need to grant additional "read" permissions to the service account that's assigned to the pods.

Here's an example RBAC configuration:

```
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: depends-on-role
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "list", "watch"]

---

kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: depends-on-role-binding
subjects:
- kind: ServiceAccount
  name: default
roleRef:
  kind: Role
  name: depends-on-role
  apiGroup: rbac.authorization.k8s.io

```

## Examples

For convenience, we've also included a few examples in the `examples/` directory. Feel free to copy-pasta from YAML files in that directory, they should provide everything you need to get going.

# Contributing

Feel free to add a PR or open an Issue if you find a bug. All contributions are welcome!

