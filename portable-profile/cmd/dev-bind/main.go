// dev-bind evaluates a Runtime Conditions Profile against the local development
// environment and writes the environment file the workload should run with.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/runtimeconditions/rc-demos/portable-profile/internal/devbind"
)

const (
	exitFailure     = 1
	exitInvalid     = 2
	exitUnsupported = 3
)

func main() {
	profilePath := flag.String("profile", "", "Runtime Conditions Profile to consume")
	todosAPIURL := flag.String("todos-api-url", "http://127.0.0.1:18081", "base URL of the local todos-api")
	redisAddr := flag.String("redis-addr", "127.0.0.1:16379", "host:port of the local redis")
	envOut := flag.String("env-out", "", "environment file to write; omit to only evaluate")
	flag.Parse()

	if *profilePath == "" {
		fail(exitFailure, errors.New("-profile is required"))
	}
	data, err := os.ReadFile(*profilePath)
	if err != nil {
		fail(exitFailure, err)
	}

	bindings, err := devbind.Bind(data, devbind.Environment{TodosAPIURL: *todosAPIURL, RedisAddr: *redisAddr})
	var invalid *devbind.InvalidProfileError
	var unsupported *devbind.UnsupportedError
	switch {
	case errors.As(err, &invalid):
		fail(exitInvalid, err)
	case errors.As(err, &unsupported):
		fail(exitUnsupported, err)
	case err != nil:
		fail(exitFailure, err)
	}

	var env strings.Builder
	fmt.Printf("dev-bind: %s\n", *profilePath)
	for _, b := range bindings {
		fmt.Printf("  condition %s (kind %s)\n", b.Condition, b.Kind)
		fmt.Printf("    support: supported\n")
		fmt.Printf("    binding: %s\n", b.Resource)
		for _, v := range b.Env {
			fmt.Printf("    env:     %s=%s\n", v.Name, v.Value)
			fmt.Fprintf(&env, "%s=%s\n", v.Name, v.Value)
		}
	}

	if *envOut != "" {
		if err := os.WriteFile(*envOut, []byte(env.String()), 0o644); err != nil {
			fail(exitFailure, err)
		}
		fmt.Printf("dev-bind: wrote %s\n", *envOut)
	}
}

func fail(code int, err error) {
	fmt.Fprintln(os.Stderr, "dev-bind:", err)
	os.Exit(code)
}
