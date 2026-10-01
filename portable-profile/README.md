# Portable Profile: one demand, two fulfillments

**Same application + same Runtime Conditions demand + different fulfillment
environments.**

This experiment tests one architectural claim: a workload can express its
Runtime Conditions once, and substantially different consumers can fulfill that
same demand differently for development and deployment without changing the
workload's semantics.

```text
apps/request-logger-http (source declarations in conditions.go)
        |
        v  go-rc-profiler (validates against extension schemas)
artifacts/request-logger-http.profile.yaml   <- the only Profile
        |
        +-----------------------------+
        v                             v
development consumer             deployment consumer
portable-profile/cmd/dev-bind    kratix ApplicationRelease resolver
        |                             |
local todos-api, local redis     platform catalog, Redis Promise
        |                             |
        v                             v
request-logger-http /demo        request-logger-http /demo
todosApi ok, cache ok            todosApi ok, cache ok
```

## The same demand, bound differently

| Condition | Development binding | Kratix binding |
| --- | --- | --- |
| `todos-api` (`api`, HTTP, `GET /todos/{id}`) | local `apps/todos-api` process | catalog API `todos-api`, contract-checked against its OpenAPI |
| `request-cache` (`cache`, key/value Redis) | local `redis-server` or `redis:7-alpine` container | `Redis` resource requested through the Redis Promise |
| Env supplied | `TODOS_API_URL`, `REDIS_URL` | `TODOS_API_URL`, `REDIS_HOST`, `REDIS_PORT` |

`request-cache` declares two configuration alternatives: `REDIS_URL`, or
`REDIS_HOST` plus `REDIS_PORT`. Each consumer chooses exactly one complete
alternative. Development takes the first satisfiable alternative in declared
order. Kratix prefers host and port because an in-cluster Redis is naturally a
Service host and port. The application accepts either, so neither choice
changes its semantics.

## What each consumer does

Both consumers follow the same visible sequence for every Condition:

```text
portable Condition -> support evaluation -> concrete binding -> runtime configuration
```

Neither consumer is a Runtime Conditions validator. Validity is established
when the profiler generates the Profile against the extension schemas. The
consumers only check the envelope defensively (`apiVersion`, `kind`, Conditions
with unique names and kinds) and then decide whether they can fulfill each
Condition. They keep four outcomes distinct:

| Outcome | Development (`dev-bind`) | Kratix (`resolver.py`) |
| --- | --- | --- |
| Profile envelope unreadable | `InvalidProfileError`, exit 2 | `ProfileError`, status `invalidProfile` |
| Valid demand this consumer does not support | `UnsupportedError`, exit 3, every Condition listed | `UnsupportedConditionError`, status `unsupportedConditions` |
| Supported, but the platform cannot satisfy it | not applicable | `ContractError`, status `validationError` (catalog incompatibility) |
| Fulfillment or runtime failure | `/demo` reports `error` | workflow or Deployment failure |

Unsupported Conditions fail before anything is started or emitted. Nothing is
silently ignored, including Conditions the Kratix resolver used to skip, such
as a non-Redis cache.

## Run it

Clone `extensions` and `go-rc-profiler` beside `rc-demos`, as described in
[`dev-container-profile/README.md`](../dev-container-profile/README.md), and use
Go 1.25 or newer.

1. Generate the Profile, or confirm the committed copy matches source:

   ```sh
   portable-profile/scripts/generate-profile.sh
   portable-profile/scripts/generate-profile.sh --check
   ```

2. Run the development demo. It needs `redis-server` or Docker:

   ```sh
   portable-profile/scripts/run-dev.sh
   ```

3. Run the deployment demo against a cluster, as in the top-level README:

   ```sh
   kratix/scripts/00-check-prereqs.sh
   kratix/scripts/01-install-kratix.sh
   kratix/scripts/02-install-promises.sh
   kratix/scripts/03-deploy-catalog-and-provider.sh
   kratix/scripts/04-deploy-application-release.sh
   kratix/scripts/05-smoke-test.sh
   ```

   `04-deploy-application-release.sh` materializes
   `kratix/generated/request-logger-application-release.yaml` by embedding the
   generated Profile, unchanged, as `spec.profile`. Nobody edits the Profile
   between the two demos.

   The `ghcr.io/runtimeconditions/*` images are not publicly pullable yet. On
   a KinD cluster, build all four images from this checkout, including the
   resolver changes, load them, and tell the scripts not to pull:

   ```sh
   kind create cluster --name rc-demos
   kratix/scripts/load-local-images.sh
   export IMAGE_PULL_POLICY=IfNotPresent
   ```

   Then run the scripts above in the same shell.

## Tests

```sh
(cd portable-profile && go test ./...)
(cd kratix/promises/application-release/pipeline && python3 -m unittest -v test_resolver)  # needs PyYAML
```

## Deliberately out of scope

- Coder, PADE, or other richer development consumers.
- A shared adapter library, result schema, support catalog format, or provenance.
- Profile lifecycle ownership. Here the Profile is a committed, generated
  artifact guarded by `generate-profile.sh --check`.

## Questions this exposes

- **Who owns support tables?** Each consumer hard-codes the Condition kinds and
  interfaces it can fulfill. That is honest, but nothing yet describes a
  consumer's capabilities outside its code.
- **Contract checking is asymmetric.** Kratix checks the API Condition against
  the published OpenAPI, while development trusts the local `todos-api`.
  Should a development consumer check contracts too, or is that a
  platform-only concern?
- **Alternative preference is consumer policy.** Both choices are valid, but
  nothing in the Profile or the extensions says how a consumer should choose.
- **Profile envelope versus validity.** Consumers repeat a small envelope check.
  Should they also confirm the Profile was validated, for example through
  provenance, rather than trusting the pipeline?
- **Error categories.** The four outcomes above are distinct today only by
  convention, in two languages.
