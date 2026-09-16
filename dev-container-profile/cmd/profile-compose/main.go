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
		fmt.Fprintln(os.Stderr, "profile-compose:", err)
		os.Exit(1)
	}
}

func run() error {
	applicationProfilePath := flag.String("application-profile", "artifacts/application.profile.yaml", "completed application Runtime Conditions Profile")
	wrapperAdditionsPath := flag.String("wrapper-additions", "examples/dev-container.conditions.yaml", "wrapper-owned Condition additions")
	recipePath := flag.String("recipe", "examples/dev-container.compose.yaml", "composition recipe")
	outPath := flag.String("out", "artifacts/dev-container.profile.yaml", "composed Runtime Conditions Profile output")
	provenanceOutPath := flag.String("provenance-out", "artifacts/dev-container.provenance.yaml", "composition provenance output")
	flag.Parse()

	var application compose.Profile
	if err := readYAML(*applicationProfilePath, &application); err != nil {
		return err
	}
	var wrapperAdditions compose.ConditionSet
	if err := readYAML(*wrapperAdditionsPath, &wrapperAdditions); err != nil {
		return err
	}
	var recipe compose.Recipe
	if err := readYAML(*recipePath, &recipe); err != nil {
		return err
	}

	profile, provenance, err := compose.Compose(application, wrapperAdditions, recipe, compose.SourcePaths{
		ApplicationProfile: *applicationProfilePath,
		WrapperAdditions:   *wrapperAdditionsPath,
		Recipe:             *recipePath,
	})
	if err != nil {
		return err
	}

	if err := writeYAML(*outPath, profile); err != nil {
		return err
	}
	if err := writeYAML(*provenanceOutPath, provenance); err != nil {
		return err
	}

	fmt.Println("composed profile:", *outPath)
	fmt.Println("composition provenance:", *provenanceOutPath)
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
