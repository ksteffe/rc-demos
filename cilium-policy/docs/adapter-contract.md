# Adapter implementation contract

This document defines the small, testable slice left for the contributor.

## Inputs

- A Runtime Conditions Profile path
- A Backstage base URL
- A Kubernetes context
- An output path

The adapter may use the local Kubernetes credentials for the proof of concept.
Backstage remains the source of catalog identity and provenance; Kubernetes is
the source of live Service names, namespaces, and ports.

## Resolution algorithm

### 1. Read

Parse the profile and retain:

- workload identity
- condition kind and interface
- required HTTP method/path pairs
- datastore or cache engine

Match workload.uri to a Component's backstage.io/source-location annotation
(after removing Backstage's url: prefix). Use the matched Component's
Kubernetes ID as the source endpoint selector. Do not derive Kubernetes names
from the profile's workload or condition names.

### 2. Match

For each API condition:

1. List Backstage API entities with spec.type equal to openapi.
2. Parse each entity's resolved spec.definition.
3. Keep an API only when its OpenAPI Paths Object contains every requested
   method/path pair.
4. Fail clearly if zero or more than one API remains.

For datastore and cache conditions, match the platform's existing Resource
taxonomy:

- Backstage spec.type (database or cache)
- demo.platform.io/interface
- demo.platform.io/engine

Condition names are human-readable labels and are not matching keys.
The demo.platform.io annotations model platform-owned catalog metadata; they
were not derived from the profile.

### 3. Follow provenance

For an API, follow its apiProvidedBy relation to the provider Component. Read
the Component's backstage.io/kubernetes-id and
backstage.io/kubernetes-namespace annotations.

For a Resource, read the equivalent annotations directly.

Use those annotations through Backstage's Kubernetes integration, or query the
same live cluster with the provided kube-context, to find exactly one Service
carrying the Kubernetes ID. Record the actual Service name, namespace, and
port. Fail rather than guessing when the lookup is empty or ambiguous.

The test fixtures intentionally resolve as follows:

| Profile capability | Catalog entity | Provider | Live Service |
| --- | --- | --- | --- |
| GET availability + POST reservation | inventory-api | inventory-service | stock-provider-v2 |
| POST fulfillment task | fulfillment-api | fulfillment-service | dispatch-planner-v1 |
| relational Postgres | resource-database | Resource itself | resource-records-v1 |
| key/value Redis | availability-cache | Resource itself | availability-cache-v1 |

This table is an acceptance oracle, not permission to hard-code the rightmost
column.

### 4. Render

Render one CiliumNetworkPolicy in the applications namespace:

- endpointSelector matches
  backstage.io/kubernetes-id=request-coordinator
- each destination is a discovered toServices entry
- each destination includes its discovered TCP port
- HTTP destinations include rules.http entries
- Postgres and Redis stop at L3/L4 and have no HTTP rules

Convert OpenAPI path templates to anchored Cilium-compatible regular
expressions. For example:

| OpenAPI path | Cilium HTTP path |
| --- | --- |
| /items/{itemId}/availability | ^/items/[^/]+/availability$ |
| /reservations | ^/reservations$ |

Keep rules for different Services in separate egress entries. Cilium combines
all fields in an egress entry, so merging unrelated destinations and L7 rules
can grant the wrong operation to the wrong service.

## Non-goals for milestone one

- Installing or configuring Backstage, KinD, Cilium, Postgres, Redis, or NATS
- Generating workload configuration values or secrets
- NATS-aware policy generation
- External FQDN and TLS/SNI policy
- Choosing among multiple deployments of one provider
- Mutating the source Deployment
- Building a long-running controller

## Acceptance sequence

1. Apply the supplied DNS-only baseline policy.
2. Confirm scripts/05-smoke-test.sh blocked succeeds.
3. Render and apply the adapter output.
4. Confirm scripts/05-smoke-test.sh allowed succeeds.
5. Confirm scripts/06-check-l7-boundaries.sh succeeds.
6. Rename a Condition and rerun; output targets must remain unchanged.
