package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/runtimeconditions/go-rc-profiler/extensioncheck"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "profile-validate:", err)
		os.Exit(1)
	}
}

func run() error {
	profilePath := flag.String("profile", "", "Runtime Conditions Profile to validate")
	extensionsRoot := flag.String("extensions-root", "../../extensions", "root containing Runtime Conditions extension definitions")
	flag.Parse()

	if *profilePath == "" {
		return fmt.Errorf("-profile is required")
	}
	data, err := os.ReadFile(*profilePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", *profilePath, err)
	}
	if err := extensioncheck.ValidateProfileYAML(data, extensioncheck.ProfileOptions{
		CatalogRoots: []string{*extensionsRoot},
	}); err != nil {
		return fmt.Errorf("validate %s: %w", *profilePath, err)
	}

	fmt.Println("validated profile:", *profilePath)
	return nil
}
