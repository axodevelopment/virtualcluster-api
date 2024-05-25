# steps to set

docker buildx build --platform linux/amd64 -t docker.io/axodevelopment/virtualcluster-api:v1.0.a --push .

oc apply -f serviceaccount.yaml
oc apply -f clusterrole.yaml
oc apply -f clusterrolebinding.yaml
oc apply -f deployment.yaml
