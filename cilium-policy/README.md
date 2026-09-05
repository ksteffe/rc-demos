# RCProfile to Cilium policy lab

This directory is the ready-made environment for proving an RCProfile adapter
that discovers platform implementations through Backstage and renders
CiliumNetworkPolicy resources.

The contributor is **not** expected to build the cluster, application, backing
services, API catalog, or baseline network policy. Their work belongs in the
adapter directory.

## What the demo does

The included application coordinates fulfillment of community resource
requests:

1. A user submits a request for an item.
2. request-coordinator asks stock-provider-v2 whether inventory is available
   and reserves it.
3. It asks dispatch-planner-v1 to create a durable fulfillment task.
4. Request state is stored in Postgres and cached in Redis.
5. A best-effort event is published to NATS for the notification worker.

The RCProfile describes the capabilities the coordinator needs. It does not
contain the names stock-provider-v2, dispatch-planner-v1,
resource-records-v1, or availability-cache-v1. Those environment-specific
names are discovered from independently maintained Backstage entities and
their Kubernetes provenance.

## Prerequisites

- Docker
- KinD
- Helm 3
- kubectl
- curl

Run the setup from this directory:

~~~bash
./scripts/00-check-prerequisites.sh
./scripts/01-create-cluster.sh
./scripts/02-install-platform.sh
./scripts/03-build-and-deploy-apps.sh
./scripts/04-apply-baseline-policy.sh
~~~

These scripts only wrap the Helm and kubectl commands documented in
docs/manual-install.md.

Open the catalog:

~~~bash
kubectl -n backstage port-forward svc/backstage 7007:7007
~~~

Then browse to <http://localhost:7007/catalog>.

Before an adapter-generated allow policy is applied, the end-to-end request is
expected to fail because the application namespace has default-deny egress:

~~~bash
./scripts/05-smoke-test.sh blocked
~~~

After the contributor renders and applies a policy:

~~~bash
kubectl apply -f ./out/request-coordinator.cnp.yaml
./scripts/05-smoke-test.sh allowed
~~~

## Directory map

| Path | Ownership | Purpose |
| --- | --- | --- |
| adapter/ | Contributor | The only implementation exercise |
| apps/resource-demo/ | Project | Runnable microservice source and image |
| backstage/ | Project | API definitions, catalog entities, and cluster provenance |
| cluster/ | Project | KinD and namespace definitions |
| helm/ | Project | Version-pinned Cilium, Backstage, and NATS values |
| manifests/ | Project | Applications, data services, RBAC, and baseline CNP |
| profiles/ | Project | Adapter input fixtures |
| scripts/ | Project | Reproducible setup and acceptance checks |

## Initial adapter boundary

For the first milestone, implement the four conditions in
profiles/request-coordinator.yaml:

1. Read the API conditions and their required method/path pairs.
2. Query Backstage for API entities whose OpenAPI documents satisfy all
   required operations.
3. Follow Backstage's apiProvidedBy relation to a Component.
4. Use that Component's Kubernetes annotations to resolve the actual Service.
5. Render one namespaced CiliumNetworkPolicy selecting
   backstage.io/kubernetes-id=request-coordinator, with toServices,
   destination ports, and HTTP method/path rules.
6. Resolve the Postgres and Redis conditions through the supplied Backstage
   Resource metadata and render their L3/L4 rules in the same policy.

NATS and external endpoints are supplied as follow-on fixtures; they do not
need to be supported in the first milestone.

## Deliberate constraints

- Every provider Component maps to exactly one Kubernetes Service.
- Internal API traffic is plaintext HTTP so Cilium can enforce L7 rules.
- All application egress is default-denied. Small platform-owned policies keep
  the providers' existing Postgres and NATS plumbing healthy; they do not grant
  the profiled coordinator any application access.
- Kubernetes object names intentionally do not match RCProfile condition names.
- Demo credentials are local-only and must not be reused elsewhere.
