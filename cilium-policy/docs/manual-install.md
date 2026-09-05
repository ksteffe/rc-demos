# Manual installation commands

The scripts in ../scripts run these commands with error checking.

~~~bash
source ./versions.env
CONTAINER_ENGINE=podman  # or docker

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
kubectl apply -f ./manifests/dependencies.yaml
kubectl apply -k ./backstage
kubectl apply -f ./manifests/backstage-rbac.yaml

helm upgrade --install event-bus nats/nats \
  --namespace dependencies \
  --version "${NATS_CHART_VERSION}" \
  --values ./helm/nats-values.yaml \
  --wait

helm upgrade --install backstage backstage/backstage \
  --namespace backstage \
  --version "${BACKSTAGE_CHART_VERSION}" \
  --values ./helm/backstage-values.yaml \
  --wait

"${CONTAINER_ENGINE}" build -t localhost/rc-resource-demo:local ./apps/resource-demo
"${CONTAINER_ENGINE}" save localhost/rc-resource-demo:local -o /tmp/rc-resource-demo.tar
kind load image-archive /tmp/rc-resource-demo.tar --name "${KIND_CLUSTER_NAME}"
rm -f /tmp/rc-resource-demo.tar

kubectl apply -f ./manifests/applications.yaml
kubectl apply -f ./manifests/provider-egress.yaml
kubectl apply -f ./manifests/default-deny-egress.yaml
~~~

The RBAC binding is applied before the Backstage release so the service account
can read cluster resources from the moment the pod starts.

The image is loaded from an archive because `kind load docker-image` shells out
to the `docker` CLI, which is absent on Podman-only hosts.

The final kubectl apply activates egress isolation for application workloads.
DNS remains allowed so the adapter can prove the difference between name
resolution and policy-authorized reachability.
