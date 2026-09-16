# Initial profile composition experiment

This experiment implements an explicit multi-phase path from the project brief:

```text
application source
  -> application.profiler.yaml
  + application-owned additions
  -> application.profile.yaml
  -> validate
  + wrapper-owned additions
  + target composition recipe
  -> dev-container.profile.yaml
  -> validate
  -> dev-container.provenance.yaml
```

The implementation is intentionally narrow. Its purpose is to make the phase
boundaries concrete enough to learn from before deciding whether any part
belongs in reusable Runtime Conditions tooling.

## Choices in this pass

### Materialize the application-owned view

Profiler output is not treated as the final application Profile. The completion
phase adds the manually declared Google Analytics requirement and writes a
standalone `application.profile.yaml` while preserving the application's own
metadata and workload identity.

The wrapper-owned GitHub requirement is introduced only after that boundary, so
the application Profile contains application requirements and does not acquire
the dev-container's source-control requirement.

### Preserve extension-defined Condition data

Completion and composition treat each Condition as an opaque YAML mapping. They
read only `name` for collision detection and provenance. They do not model HTTP
operations, Google Analytics events, source-control access, or any other
extension-defined interface shape.

That keeps these stages independent of extension vocabulary: new Condition
kinds can pass through without changing the composer.

### Additive composition with exact extension de-duplication

Extensions are merged in first-seen order and de-duplicated by exact URI.
Conditions are appended in source order.

Condition names are treated as identities for this experiment. A duplicate name
across inputs is an error rather than an implicit override or merge. This is
deliberately conservative until Runtime Conditions defines stronger identity or
semantic merge rules.

### Validate materialized Profiles at phase boundaries

Both `application.profile.yaml` and `dev-container.profile.yaml` are validated
against the vocabulary and JSON Schemas of their declared extensions using the
existing `go-rc-profiler/extensioncheck` profile validator. Validation remains a
separate pipeline stage rather than teaching the composer extension semantics.

CI publishes the application Profile only after application validation succeeds,
and publishes the dev-container artifacts only after dev-container validation
succeeds.

### Keep workload identity separate from inherited Conditions

The completed application Profile retains the application's workload identity.
The dev-container Profile always takes `metadata` and `workload` from the
composition recipe; it does not inherit the application workload identity.
Application Conditions are copied into the new dev-container workload Profile as
requirements that the wrapper needs while building or running the application.

### Keep provenance outside the Profile

The generated `dev-container.profile.yaml` remains an ordinary
`RuntimeConditionsProfile`. Origin information is emitted separately as
`dev-container.provenance.yaml`.

At the dev-container boundary, the completed application Profile is one source
and the wrapper-owned additions are another. This keeps ownership visible
without adding composition-only fields to the Runtime Conditions Profile format.
Whether provenance should later chain through the application Profile to its
profiler/manual inputs remains open.

## What this pass does not decide

This experiment does not yet answer:

- whether Condition names are sufficient stable identity long term;
- whether two Conditions with different names can be semantically equivalent;
- whether provenance should eventually have a standard schema or chained source history;
- whether extension versions should be resolved and locked before composition;
- whether removals or semantic changes should trigger recomposition;
- how a failed composition should preserve or publish a last-known-good Profile.

Those questions are easier to evaluate now that the materialized Profiles and
validation boundaries exist as explicit CI artifacts.
