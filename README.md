# Runtime Conditions Demos

Runtime Conditions is currently seeking adoption by an established parent
project. The repositories in this organization are split for hands-on usability,
review, demos, and implementation feedback. They are not intended to present
Runtime Conditions as a standalone foundation or competing project.

Start here: https://runtimeconditions.github.io/

This tree contains runnable examples and downstream adapter assets.

## Layout

- `apps/request-logger-http/` - Go workload that imports first-party declaration packages and demonstrates explicit profile declarations.
- `apps/request-logger-http-java/` - Java workload with matching explicit declarations for the same Conditions as the Go request logger.
- `apps/todos-api/` - simple provider API used by the request logger demo.
- `dev-container-profile/` - development scaffold and acceptance contract for a build-time Profile composition demo. Implementers should start with [`dev-container-profile/README.md`](dev-container-profile/README.md).
- `catalog/apis/` - OpenAPI and catalog files used by the adapter demo.
- `kratix/` - Kratix Promise and adapter assets for downstream fulfillment demos.
- `kratix/manifests/` - static Kubernetes and Kratix manifests applied by the demo scripts.
- `cilium-policy/` - a complete KinD, Cilium, Backstage, and microservice lab for
  building an RCProfile-to-CiliumNetworkPolicy adapter.

## Generate the Request Logger Profile

From the repository root:

```sh
cd ../go-rc-profiler
go run . \
  -dir ../rc-demos/apps/request-logger-http \
  -workload-uri github.com/runtimeconditions/rc-demos/apps/request-logger-http \
  -name request-logger-http \
  -workload-version dev
```

The request logger is its own Go module:

```sh
cd apps/request-logger-http
go test ./...
```

## Generate the Java Request Logger Profile

From the repository root:

```sh
cd ../java-rc-profiler
mvn -q package

java -jar target/runtimeconditions-java-profiler-0.1.0-SNAPSHOT.jar generate \
  --project ../rc-demos/apps/request-logger-http-java \
  --classpath ../extensions/common-integrations/java:../extensions/env-configuration/java \
  --name request-logger-http \
  --workload-uri github.com/runtimeconditions/rc-demos/apps/request-logger-http-java \
  --workload-version dev
```

The Java demo is a Maven project:

```sh
cd apps/request-logger-http-java
mvn -q package
```

## Published Demo Images

The image names used by the manifests are intended to live under
`ghcr.io/runtimeconditions/` once publishing workflows are split out.

- `redis-pipeline`
- `application-release-pipeline`
- `todos-api`
- `request-logger`

## Run the Kratix Demo

From the repository root:

```sh
kratix/scripts/00-check-prereqs.sh
kratix/scripts/01-install-kratix.sh
kratix/scripts/02-install-promises.sh
kratix/scripts/03-deploy-catalog-and-provider.sh
kratix/scripts/04-deploy-application-release.sh
kratix/scripts/05-smoke-test.sh
```

To run the contract failure path:

```sh
kratix/scripts/06-demo-breaking-change.sh
```
