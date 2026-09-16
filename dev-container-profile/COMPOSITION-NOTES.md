# Initial profile composition experiment

This experiment implements the smallest explicit composition path from the project brief:

```text
application source
  -> application.profiler.yaml
  + application-owned additions
  + wrapper-owned additions
  + target composition recipe
  -> dev-container.profile.yaml
  -> dev-container.provenance.yaml
```

The implementation is intentionally narrow. Its purpose is to make the phase boundary concrete enough to learn from before deciding whether any part belongs in reusable Runtime Conditions tooling.

## Choices in this pass

### Preserve extension-defined Condition data

The composer treats each Condition as an opaque YAML mapping. It reads only `name` for collision detection and provenance. It does not model HTTP operations, Google Analytics events, source-control access, or any other extension-defined interface shape.

That keeps composition independent of extension vocabulary: new Condition kinds can pass through without changing the composer.

### Additive composition with exact extension de-duplication

Extensions are merged in first-seen order and de-duplicated by exact URI. Conditions are appended in source order.

Condition names are treated as identities for this experiment. A duplicate name across inputs is an error rather than an implicit override or merge. This is deliberately conservative until Runtime Conditions defines stronger identity or semantic merge rules.

### Keep workload identity separate from inherited Conditions

The composed Profile always takes `metadata` and `workload` from the composition recipe. It does not inherit the application workload identity. Application Conditions are copied into the new dev-container workload Profile as requirements that the wrapper needs while building or running the application.

### Keep provenance outside the Profile

The generated `dev-container.profile.yaml` remains an ordinary `RuntimeConditionsProfile`. Origin information is emitted separately as `dev-container.provenance.yaml`, which records which source contributed each named Condition.

This avoids adding composition-only fields to the Runtime Conditions Profile format while still making attribution inspectable.

## What this pass does not decide

This experiment does not yet answer:

- whether Condition names are sufficient stable identity long term;
- whether two Conditions with different names can be semantically equivalent;
- whether provenance should eventually have a standard schema;
- whether extension versions should be resolved and locked before composition;
- whether removals or semantic changes should trigger recomposition;
- whether application completion should be materialized as its own intermediate Profile;
- how a failed composition should preserve or publish a last-known-good Profile.

Those questions are easier to evaluate once the explicit artifacts exist and can be passed through CI.
