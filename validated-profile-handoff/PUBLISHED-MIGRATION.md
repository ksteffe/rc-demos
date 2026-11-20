# Published distribution migration: official rc-cli validation

This branch follows merged
[rc-demos PR #5](https://github.com/runtimeconditions/rc-demos/pull/5)
and Colin's feedback about published tool and catalog usage.

## Implementation (2026-11-19)

The official [rc-cli v0.6.0 release](https://github.com/runtimeconditions/rc-cli/releases/tag/v0.6.0)
(published November 18, 2026) embeds the Go profiler **v0.8.0**
and core schema **0.4.0**. The repository installs that CLI from its pinned
release archives, checking the published SHA-256 on every invocation. Local
Homebrew/apt/dnf installs are supported for exploration but not used as a
floating CI dependency.

`portable-profile/scripts/generate-profile.sh` now executes
`rc profile generate`. The new CLI's canonical serialization writes
`https://runtimeconditions.io/extensions/common-integrations:v1alpha1`
and `https://runtimeconditions.io/extensions/env-configuration:v1alpha1`
as short, versioned references. The committed Profile is regenerated in that
format, with its original bytes handed to consumers unchanged.

The trusted producer expands those short references into complete HTTPS
definition URLs **for resolution and evidence identity only**. It captures
every retrieved definition once, including dependencies, and writes the exact
response bytes to a private snapshot. It then invokes
`rc profile validate - --catalog <snapshot>`, passing the unmodified Profile
through stdin. This is the independently implemented CLI validator—not
`go-rc-profiler/extensioncheck`—and it correctly matches complete definition
IDs with optional `metadata.version` annotations. `--catalog` forbids
additional live fetches during validation.

Only a successful CLI validation allows the producer to sign evidence.
The signature covers the exact Profile digest, full extension ID/digests,
validator identity and executable SHA-256, and producer identity. The
downstream consumer still links no RC semantic validator and still performs
support evaluation only after signature and byte-digest verification.

No unmerged upstream Go profiler patch, local checkout substitution, modified
published definitions, or rewritten Profile artifact is required.

## Remaining publication blocker

GitHub Actions on November 19, 2026 returned **HTTP 404** from
`https://runtimeconditions.io/extensions/common-integrations/v1alpha1/runtimeconditions.extension.yaml`
and the corresponding published catalog route could not be used for the
required live HTTPS acceptance. This occurs during snapshot retrieval,
before the CLI gets to validate. Local captured-publication fixtures are
useful for testing the new CLI's compatibility, but **do not establish**
that the upstream catalog is live or that the five live handoff cases pass.

Do not replace those URLs with invented aliases or silently fall back to
local copies. The required next step is to restore the official published
definition URLs (or adopt newly published authoritative identifiers through
the bindings, if that is the upstream direction), then rerun all five live
acceptance cases and real KinD/Kratix. Keep failures visible and the PR open.

## Reproduce

```sh
export RC_CLI="$(sh portable-profile/scripts/install-rc.sh)"
portable-profile/scripts/generate-profile.sh --check
(cd validated-profile-handoff && go test -v ./internal/producer)
(cd validated-profile-handoff && go test -v ./...)
validated-profile-handoff/scripts/run.sh
```

The offline producer regression uses the committed, unmodified published
definition fixtures. Required acceptance and Kratix runs use HTTPS by
default and must not claim success when publication is unavailable.

## Validation evidence and limitations

The original evidence format and signing flow are unchanged. Evidence covers the
exact Profile bytes (including the generated header) and SHA-256 digests of every
resolved extension response; the signature covers the exact JSON evidence bytes.
Complete bundles are atomically renamed into place only after strict validation;
keys remain outside published output. Unavailable resources, invalid definitions,
conflicts, and validation errors produce no signed artifacts and no fallback.

This remains experimental implementation evidence, not a proposed standardized
evidence format or production-grade signing infrastructure. The original trust,
replay, publisher, and deployment limitations remain.

## Local verification (2026-10-21)

- Workload Go tests, portable consumer Go tests, and Profile drift check pass.
- Producer catalog tests pass (HTTP fixture failures, closure, exact bytes,
  conflicts/cycles, response limits, explicit HTTPS policy, and the incompatibility
  reproducer). The consumer dependency boundary test passes.
- Existing Python ApplicationRelease resolver suite: 24 tests pass, including
  canonical Redis/Deployment/Service output and optional-name handling.
- Go vet passes for the workload, portable consumer, and handoff module.
- Five changed/new shell scripts pass syntax checks; both workflow YAML files
  parse and retain `contents: read`. ShellCheck is not installed locally.
- Full handoff Go acceptance suite, Python platform adapter setup, and the real
  runner fail at the published definition HTTP 404 described above.
- Docker, KinD, and kubectl are unavailable locally. The cluster job is retained
  in CI and depends on the handoff job; it cannot run successfully until the
  producer blocker is fixed. No cluster acceptance pass is claimed.
- The live development demo also needs Redis or Docker, unavailable locally;
  its existing CI step is retained.
