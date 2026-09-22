# Dev-container Profile exploration

**Start here.** This is a workspace for exploring how one workload's Runtime
Conditions can contribute to the Profile of another workload—for example, how
an application's requirements can be included in the Profile of the development
container used to build it.

The project brief in [`PROJECT-BRIEF.md`](PROJECT-BRIEF.md) describes the
scenario, the open questions, and several possible directions without requiring
a particular architecture. This branch adds one deliberately small composition
experiment so those questions can be tested against explicit artifacts.

## What is ready to use

| Area | What it provides |
| --- | --- |
| `.devcontainer/` | A Go 1.25 environment with the sibling repositories mounted together |
| `app/` | A tested web application plus a small content API |
| `app/profile/conditions.go` | A declaration the existing Go profiler can discover |
| `scripts/generate-application-profile.sh` | Profiling phase that emits `artifacts/application.profiler.yaml` |
| `cmd/profile-complete/` | Materializes the application-owned Profile from profiler output plus manual additions |
| `cmd/profile-compose/` | Additive dev-container composer that preserves extension-defined Conditions as opaque YAML |
| `cmd/profile-validate/` | Validates materialized Profiles against their declared extension schemas |
| `scripts/compose-dev-container-profile.sh` | End-to-end profiling, completion, validation, and composition helper |
| `COMPOSITION-NOTES.md` | Design choices, non-goals, and questions exposed by the experiment |
| `scripts/smoke-test.sh` | A real HTTP check of the target application |
| `examples/` | Possible manual GA, wrapper, and composition inputs—not required formats |
| `reference/` | Illustrations of possible application and dev-container outputs—not golden files |
| sibling `extensions/` | Extension vocabulary and validation schemas used by the demo |
| `.github/workflows/dev-container-profile.yaml` | Independent pipeline stages that pass real artifacts through profiling, completion, and composition |

Google Analytics is represented manually for now. Discovering it automatically
from application code is a possible later experiment, not a prerequisite.

## Get oriented

This demo uses three separate Git repositories:

- [`runtimeconditions/rc-demos`](https://github.com/runtimeconditions/rc-demos)
- [`runtimeconditions/extensions`](https://github.com/runtimeconditions/extensions)
- [`runtimeconditions/go-rc-profiler`](https://github.com/runtimeconditions/go-rc-profiler)

Clone all three repositories into the same parent directory, using these
directory names:

```sh
mkdir runtimeconditions
cd runtimeconditions
git clone https://github.com/runtimeconditions/rc-demos.git
git clone https://github.com/runtimeconditions/extensions.git
git clone https://github.com/runtimeconditions/go-rc-profiler.git
```

The resulting local workspace must have this sibling layout:

```text
runtimeconditions/
├── extensions/
├── go-rc-profiler/
└── rc-demos/
    └── dev-container-profile/
```

Then open `rc-demos/dev-container-profile/` in VS Code and choose **Reopen in
Container**. The development-container configuration mounts the shared parent
directory so it can access the separately cloned `extensions` and
`go-rc-profiler` repositories through the relative paths used by the demo.

Without VS Code Dev Containers, install Go 1.25, Git, Python 3, and curl. The
three repositories must still be checked out with the same sibling layout.

Run the supplied pieces independently:

```sh
go test ./...
./scripts/smoke-test.sh
./scripts/generate-application-profile.sh
```

Run the end-to-end composition experiment with:

```sh
sh ./scripts/compose-dev-container-profile.sh
```

That produces and validates:

```text
artifacts/application.profiler.yaml
artifacts/application.profile.yaml
artifacts/dev-container.profile.yaml
artifacts/dev-container.provenance.yaml
```

The profiler output is first completed with the manual application-owned Google
Analytics Condition while preserving the application's workload identity. That
standalone application Profile is validated against its declared extension
schemas before it becomes an input to the dev-container composition phase.

The dev-container composer then combines the completed application Profile with
the wrapper-owned GitHub Condition and assigns the workload identity from the
composition recipe. Extension URIs are de-duplicated exactly, duplicate
Condition names fail rather than overwrite one another, and the finished
dev-container Profile is validated before the pipeline publishes it.
Composition provenance stays in a separate artifact so the materialized Profile
remains an ordinary `RuntimeConditionsProfile`.

Read [`COMPOSITION-NOTES.md`](COMPOSITION-NOTES.md) for the choices this pass
makes deliberately and the questions it leaves open.
