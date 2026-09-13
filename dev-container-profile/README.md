# Dev-container Profile exploration

**Start here.** This is a workspace for exploring how one workload's Runtime
Conditions can contribute to the Profile of another workload—for example, how
an application's requirements can be included in the Profile of the development
container used to build it.

The composition solution has intentionally not been designed or implemented.
The project brief in [`PROJECT-BRIEF.md`](PROJECT-BRIEF.md) describes the
scenario, the open questions, and several possible directions without requiring
a particular architecture.

## What is ready to use

| Area | What it provides |
| --- | --- |
| `.devcontainer/` | A Go 1.25 environment with the sibling repositories mounted together |
| `app/` | A tested web application plus a small content API |
| `app/profile/conditions.go` | A declaration the existing Go profiler can discover |
| `scripts/generate-application-profile.sh` | One profiling phase that emits `artifacts/application.profiler.yaml` |
| `scripts/smoke-test.sh` | A real HTTP check of the target application |
| `examples/` | Possible manual GA, wrapper, and composition inputs—not required formats |
| `reference/` | Illustrations of possible application and dev-container outputs—not golden files |
| sibling `extensions/` | Validating Google Analytics and source-control vocabulary |
| `.github/workflows/dev-container-profile.yaml` | Independent foundation stages that a contributor can extend |

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

After those work, read [`PROJECT-BRIEF.md`](PROJECT-BRIEF.md) and choose the
part of the pipeline you would like to explore first.

There is no single expected executable, package layout, intermediate format, or
watching mechanism. A useful contribution may begin as a design note, a small
pipeline stage, an experiment comparing representations, or a complete
end-to-end demo.
