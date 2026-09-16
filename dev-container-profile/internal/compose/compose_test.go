package compose

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestComposeBuildsTargetProfileAndProvenance(t *testing.T) {
	application := mustDecode[Profile](t, `
apiVersion: runtimeconditions.io/v1alpha1
kind: RuntimeConditionsProfile
metadata:
  name: web-profile-demo
workload:
  uri: https://example.test/app
  version: demo
extensions:
  - https://example.test/extensions/common.yaml
conditions:
  - name: content-api
    kind: api
    interface:
      type: http
      operations:
        - method: GET
          path: /message
`)
	applicationAdditions := mustDecode[ConditionSet](t, `
extensions:
  - https://example.test/extensions/analytics.yaml
conditions:
  - name: site-analytics
    kind: google.analytics
    interface:
      type: web
      events:
        - page_view
`)
	wrapperAdditions := mustDecode[ConditionSet](t, `
extensions:
  - https://example.test/extensions/source-control.yaml
conditions:
  - name: application-source
    kind: source_control
    interface:
      type: git
      provider: github
      access:
        - fetch
        - pull
        - push
`)
	recipe := mustDecode[Recipe](t, `
apiVersion: tooling.runtimeconditions.io/v1alpha1
kind: ProfileCompositionRecipe
metadata:
  name: web-demo-dev-container
target:
  metadata:
    name: web-demo-dev-container
    labels:
      lifecycle.example.com/stage: development
  workload:
    uri: https://example.test/app#dev-container
    version: demo
`)

	profile, provenance, err := Compose(application, applicationAdditions, wrapperAdditions, recipe, SourcePaths{
		ApplicationProfile:   "artifacts/application.profiler.yaml",
		ApplicationAdditions: "examples/application.conditions.yaml",
		WrapperAdditions:     "examples/dev-container.conditions.yaml",
		Recipe:               "examples/dev-container.compose.yaml",
	})
	if err != nil {
		t.Fatalf("Compose() error = %v", err)
	}

	if profile.Metadata.Name != "web-demo-dev-container" {
		t.Fatalf("target name = %q", profile.Metadata.Name)
	}
	if got := profile.Metadata.Labels["tooling.runtimeconditions.io/generated-by"]; got != "profile-compose" {
		t.Fatalf("generated-by label = %q", got)
	}
	if len(profile.Extensions) != 3 {
		t.Fatalf("extensions = %d, want 3", len(profile.Extensions))
	}
	if len(profile.Conditions) != 3 {
		t.Fatalf("conditions = %d, want 3", len(profile.Conditions))
	}

	wantNames := []string{"content-api", "site-analytics", "application-source"}
	for i := range profile.Conditions {
		name, err := conditionName(&profile.Conditions[i])
		if err != nil {
			t.Fatalf("condition %d: %v", i, err)
		}
		if name != wantNames[i] {
			t.Fatalf("condition %d name = %q, want %q", i, name, wantNames[i])
		}
	}

	if provenance.Kind != ProvenanceKind {
		t.Fatalf("provenance kind = %q", provenance.Kind)
	}
	if len(provenance.Sources) != 3 {
		t.Fatalf("provenance sources = %d, want 3", len(provenance.Sources))
	}
	if got := provenance.Sources[2].Conditions; len(got) != 1 || got[0] != "application-source" {
		t.Fatalf("wrapper provenance conditions = %#v", got)
	}
}

func TestComposeRejectsConditionNameCollisions(t *testing.T) {
	application := mustDecode[Profile](t, `
apiVersion: runtimeconditions.io/v1alpha1
kind: RuntimeConditionsProfile
metadata:
  name: app
workload:
  uri: https://example.test/app
conditions:
  - name: duplicated
    kind: api
    interface:
      type: http
`)
	applicationAdditions := mustDecode[ConditionSet](t, `
conditions:
  - name: duplicated
    kind: google.analytics
    interface:
      type: web
`)
	recipe := mustDecode[Recipe](t, `
apiVersion: tooling.runtimeconditions.io/v1alpha1
kind: ProfileCompositionRecipe
metadata:
  name: dev
target:
  metadata:
    name: dev
  workload:
    uri: https://example.test/app#dev-container
`)

	_, _, err := Compose(application, applicationAdditions, ConditionSet{}, recipe, SourcePaths{})
	if err == nil || !strings.Contains(err.Error(), "collides") {
		t.Fatalf("Compose() error = %v, want collision error", err)
	}
}

func mustDecode[T any](t *testing.T, input string) T {
	t.Helper()
	var out T
	if err := yaml.Unmarshal([]byte(input), &out); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return out
}
