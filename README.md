# VirtualCluster API's

This project represents various APIs around the ocp-virtual project. I am initially setting this up the OpenShift dynamic plugin that will be the UI view of VirtualCluster's in openshift.

This is the operator repo
https://github.com/axodevelopment/ocp-virtualcluster

This is the dynamic plug repo
https://github.com/axodevelopment/virtualcluster-plugin

docker buildx build --platform linux/amd64 -t docker.io/axodevelopment/virtualcluster-api:v1.0.g --push .

```bash

kind: Route
apiVersion: route.openshift.io/v1
metadata:
  name: virtualcluster-api
  namespace: virtualcluster-system
  labels:
    app: virtualcluster-api
spec:
  to:
    kind: Service
    name: virtualcluster-api
  tls:
    termination: edge
  port:
    targetPort: 8080
```
