package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/runtimeconditions/rc-demos/dev-container-profile/internal/compose"
	"gopkg.in/yaml.v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "profile-complete:", err)
		os.Exit(1)
	}
}

func run() error {
	applicationProfilePath := flag.String("application-profile", "artifacts/application.profiler.yaml", "profiler-produced application Runtime Conditions Profile")
	applicationAdditionsPath := flag.String("application-additions", "examples/application.conditions.yaml", "application-owned Condition additions")
	outPath := flag.String("out", "artifacts/application.profile.yaml", "completed application Runtime Conditions Profile output")
	flag.Parse()

	var application compose.Profile
	if err := readYAML(*applicationProfilePath, &application); err != nil {
		return err
	}
	var additions compose.ConditionSet
	if err := readYAML(*applicationAdditionsPath, &additions); err != nil {
		return err
	}

	profile, err := compose.CompleteApplication(application, additions)
	if err != nil {
		return err
	}
	if err := writeYAML(*outPath, profile); err != nil {
		return err
	}

	fmt.Println("completed application profile:", *outPath)
	return nil
}

func readYAML(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

func writeYAML(path string, value any) error {
	data, err := yaml.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory for %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
