# Manual installation commands

The scripts in ../scripts run these commands with error checking.

~~~bash
source ./versions.env

kind create cluster \
  --name "${KIND_CLUSTER_NAME}" \
  --image "${KIND_NODE_IMAGE}" \
  --config ./cluster/kind.yaml

helm repo add cilium https://helm.cilium.io/
helm repo add backstage https://backstage.github.io/charts
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update

helm upgrade --install cilium cilium/cilium \
  --namespace kube-system \
  --version "${CILIUM_CHART_VERSION}" \
  --values ./helm/cilium-values.yaml \
  --wait

kubectl apply -f ./cluster/namespaces.yaml
kubectl apply -k ./backstage

helm upgrade --install backstage backstage/backstage \
  --namespace backstage \
  --version "${BACKSTAGE_CHART_VERSION}" \
  --values ./helm/backstage-values.yaml \
  --wait

kubectl apply -f ./manifests/backstage-rbac.yaml
kubectl apply -f ./manifests/dependencies.yaml

helm upgrade --install event-bus nats/nats \
  --namespace dependencies \
  --version "${NATS_CHART_VERSION}" \
  --values ./helm/nats-values.yaml \
  --wait

docker build -t rc-resource-demo:local ./apps/resource-demo
kind load docker-image rc-resource-demo:local --name "${KIND_CLUSTER_NAME}"
kubectl apply -f ./manifests/applications.yaml
kubectl apply -f ./manifests/provider-egress.yaml
kubectl apply -f ./manifests/default-deny-egress.yaml
~~~

The final kubectl apply activates egress isolation for application workloads.
DNS remains allowed so the adapter can prove the difference between name
resolution and policy-authorized reachability.
