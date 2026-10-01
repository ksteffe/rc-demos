package devbind

import (
	"errors"
	"os"
	"strings"
	"testing"
)

var testEnv = Environment{TodosAPIURL: "http://127.0.0.1:18081", RedisAddr: "127.0.0.1:16379"}

func canonicalProfile(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../../artifacts/request-logger-http.profile.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func profileWith(conditions string) []byte {
	return []byte(`apiVersion: runtimeconditions.io/v1alpha1
kind: RuntimeConditionsProfile
metadata:
  name: test
conditions:
` + conditions)
}

const todosCondition = `
  - name: todos-api
    kind: api
    interface:
      type: http
      spec:
        format: openapi
        uri: catalog://api/default/todos-api
        version: 1.0.0
    configuration:
      env:
        - property: baseUrl
          name: TODOS_API_URL
`

const redisCondition = `
  - name: request-cache
    kind: cache
    interface:
      type: key_value
      engine: redis
    configuration:
      alternatives:
        - env:
            - property: url
              name: REDIS_URL
        - env:
            - property: hostname
              name: REDIS_HOST
            - property: port
              name: REDIS_PORT
`

func envOf(bindings []Binding) map[string]string {
	out := map[string]string{}
	for _, b := range bindings {
		for _, v := range b.Env {
			out[v.Name] = v.Value
		}
	}
	return out
}

func TestCanonicalProfileBindsLocally(t *testing.T) {
	bindings, err := Bind(canonicalProfile(t), testEnv)
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(bindings))
	}
	got := envOf(bindings)
	want := map[string]string{
		"TODOS_API_URL": "http://127.0.0.1:18081",
		"REDIS_URL":     "redis://127.0.0.1:16379",
	}
	if len(got) != len(want) {
		t.Fatalf("env = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
}

func TestSupportsTodosAPICondition(t *testing.T) {
	bindings, err := Bind(profileWith(todosCondition), testEnv)
	if err != nil {
		t.Fatal(err)
	}
	if got := envOf(bindings)["TODOS_API_URL"]; got != testEnv.TodosAPIURL {
		t.Fatalf("TODOS_API_URL = %q", got)
	}
}

func TestSupportsRedisConditionWithFirstCompleteAlternative(t *testing.T) {
	bindings, err := Bind(profileWith(redisCondition), testEnv)
	if err != nil {
		t.Fatal(err)
	}
	env := envOf(bindings)
	if env["REDIS_URL"] != "redis://127.0.0.1:16379" {
		t.Fatalf("REDIS_URL = %q", env["REDIS_URL"])
	}
	if _, ok := env["REDIS_HOST"]; ok {
		t.Fatalf("expected exactly one alternative, got %v", env)
	}
}

func TestRejectsUnknownCondition(t *testing.T) {
	_, err := Bind(profileWith(todosCondition+`
  - name: site-analytics
    kind: google.analytics
    interface:
      type: web
    configuration:
      env:
        - property: measurementId
          name: GA_MEASUREMENT_ID
`), testEnv)
	var unsupported *UnsupportedError
	if !errors.As(err, &unsupported) {
		t.Fatalf("expected UnsupportedError, got %v", err)
	}
	if len(unsupported.Conditions) != 1 || unsupported.Conditions[0].Name != "site-analytics" {
		t.Fatalf("unexpected unsupported conditions: %+v", unsupported.Conditions)
	}
	if !strings.Contains(err.Error(), `"google.analytics"`) {
		t.Fatalf("error does not identify the kind: %v", err)
	}
}

func TestRejectsUnsupportedCacheEngine(t *testing.T) {
	_, err := Bind(profileWith(strings.Replace(redisCondition, "engine: redis", "engine: memcached", 1)), testEnv)
	var unsupported *UnsupportedError
	if !errors.As(err, &unsupported) || unsupported.Conditions[0].Name != "request-cache" {
		t.Fatalf("expected request-cache to be unsupported, got %v", err)
	}
}

func TestRejectsUnprovidedProperty(t *testing.T) {
	_, err := Bind(profileWith(strings.Replace(todosCondition, "property: baseUrl", "property: token", 1)), testEnv)
	var unsupported *UnsupportedError
	if !errors.As(err, &unsupported) || !strings.Contains(err.Error(), `"token"`) {
		t.Fatalf("expected unsupported property token, got %v", err)
	}
}

func TestRejectsInvalidEnvelope(t *testing.T) {
	_, err := Bind([]byte("apiVersion: v1\nkind: ConfigMap\nconditions:\n  - kind: api\n"), testEnv)
	var invalid *InvalidProfileError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected InvalidProfileError, got %v", err)
	}
	if len(invalid.Problems) != 3 {
		t.Fatalf("expected apiVersion, kind and name problems, got %v", invalid.Problems)
	}
}
