package compose

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	RuntimeConditionsAPIVersion = "runtimeconditions.io/v1alpha1"
	RuntimeConditionsKind       = "RuntimeConditionsProfile"
	RecipeAPIVersion            = "tooling.runtimeconditions.io/v1alpha1"
	RecipeKind                  = "ProfileCompositionRecipe"
	ProvenanceKind              = "ProfileCompositionProvenance"
)

type Metadata struct {
	Name        string            `yaml:"name"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type Workload struct {
	URI     string `yaml:"uri"`
	Version string `yaml:"version,omitempty"`
}

type Profile struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   Metadata    `yaml:"metadata"`
	Workload   Workload    `yaml:"workload"`
	Extensions []string    `yaml:"extensions,omitempty"`
	Conditions []yaml.Node `yaml:"conditions,omitempty"`
}

type ConditionSet struct {
	Extensions []string    `yaml:"extensions,omitempty"`
	Conditions []yaml.Node `yaml:"conditions,omitempty"`
}

type Recipe struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Target struct {
		Metadata Metadata `yaml:"metadata"`
		Workload Workload `yaml:"workload"`
	} `yaml:"target"`
}

type SourcePaths struct {
	ApplicationProfile string
	WrapperAdditions   string
	Recipe             string
}

type Provenance struct {
	APIVersion string             `yaml:"apiVersion"`
	Kind       string             `yaml:"kind"`
	Metadata   Metadata           `yaml:"metadata"`
	Target     ProvenanceTarget   `yaml:"target"`
	Sources    []ProvenanceSource `yaml:"sources"`
}

type ProvenanceTarget struct {
	Profile  string   `yaml:"profile"`
	Workload Workload `yaml:"workload"`
}

type ProvenanceSource struct {
	ID         string   `yaml:"id"`
	Type       string   `yaml:"type"`
	Path       string   `yaml:"path,omitempty"`
	Conditions []string `yaml:"conditions,omitempty"`
}

// CompleteApplication materializes the application-owned view by combining the
// profiler output with manually authored application requirements. The
// application's workload identity is preserved and wrapper-owned requirements
// are intentionally not accepted here.
func CompleteApplication(application Profile, additions ConditionSet) (Profile, error) {
	if err := validateProfileEnvelope(application, "application profile"); err != nil {
		return Profile{}, err
	}

	out := Profile{
		APIVersion: application.APIVersion,
		Kind:       application.Kind,
		Metadata:   cloneMetadata(application.Metadata),
		Workload:   application.Workload,
	}

	seenExtensions := map[string]struct{}{}
	var err error
	out.Extensions, err = mergeExtensions(seenExtensions, out.Extensions, application.Extensions, additions.Extensions)
	if err != nil {
		return Profile{}, err
	}

	seenConditions := map[string]string{}
	if _, err := appendConditions(&out.Conditions, seenConditions, "application-profile", application.Conditions); err != nil {
		return Profile{}, err
	}
	if _, err := appendConditions(&out.Conditions, seenConditions, "application-additions", additions.Conditions); err != nil {
		return Profile{}, err
	}

	return out, nil
}

// Compose materializes the dev-container view from an already completed
// application Profile plus wrapper-owned requirements and a target recipe.
func Compose(application Profile, wrapperAdditions ConditionSet, recipe Recipe, paths SourcePaths) (Profile, Provenance, error) {
	if err := validateProfileEnvelope(application, "application profile"); err != nil {
		return Profile{}, Provenance{}, err
	}
	if recipe.APIVersion != RecipeAPIVersion || recipe.Kind != RecipeKind {
		return Profile{}, Provenance{}, fmt.Errorf("composition recipe must be %s %s, got %q %q", RecipeAPIVersion, RecipeKind, recipe.APIVersion, recipe.Kind)
	}
	if strings.TrimSpace(recipe.Target.Metadata.Name) == "" {
		return Profile{}, Provenance{}, fmt.Errorf("composition target metadata.name is required")
	}
	if strings.TrimSpace(recipe.Target.Workload.URI) == "" {
		return Profile{}, Provenance{}, fmt.Errorf("composition target workload.uri is required")
	}

	metadata := cloneMetadata(recipe.Target.Metadata)
	if metadata.Labels == nil {
		metadata.Labels = map[string]string{}
	}
	metadata.Labels["tooling.runtimeconditions.io/generated-by"] = "profile-compose"

	out := Profile{
		APIVersion: RuntimeConditionsAPIVersion,
		Kind:       RuntimeConditionsKind,
		Metadata:   metadata,
		Workload:   recipe.Target.Workload,
	}

	seenExtensions := map[string]struct{}{}
	var err error
	out.Extensions, err = mergeExtensions(seenExtensions, out.Extensions, application.Extensions, wrapperAdditions.Extensions)
	if err != nil {
		return Profile{}, Provenance{}, err
	}

	seenConditions := map[string]string{}
	sources := make([]ProvenanceSource, 0, 2)
	appendSource := func(id, sourceType, path string, conditions []yaml.Node) error {
		names, err := appendConditions(&out.Conditions, seenConditions, id, conditions)
		if err != nil {
			return err
		}
		sources = append(sources, ProvenanceSource{
			ID:         id,
			Type:       sourceType,
			Path:       path,
			Conditions: names,
		})
		return nil
	}

	if err := appendSource("application-profile", "profile", paths.ApplicationProfile, application.Conditions); err != nil {
		return Profile{}, Provenance{}, err
	}
	if err := appendSource("wrapper-additions", "condition-set", paths.WrapperAdditions, wrapperAdditions.Conditions); err != nil {
		return Profile{}, Provenance{}, err
	}

	provenance := Provenance{
		APIVersion: RecipeAPIVersion,
		Kind:       ProvenanceKind,
		Metadata:   Metadata{Name: recipe.Target.Metadata.Name},
		Target: ProvenanceTarget{
			Profile:  recipe.Target.Metadata.Name,
			Workload: recipe.Target.Workload,
		},
		Sources: sources,
	}

	return out, provenance, nil
}

func validateProfileEnvelope(profile Profile, description string) error {
	if profile.APIVersion != RuntimeConditionsAPIVersion || profile.Kind != RuntimeConditionsKind {
		return fmt.Errorf("%s must be %s %s, got %q %q", description, RuntimeConditionsAPIVersion, RuntimeConditionsKind, profile.APIVersion, profile.Kind)
	}
	return nil
}

func mergeExtensions(seen map[string]struct{}, dst []string, groups ...[]string) ([]string, error) {
	for _, group := range groups {
		for _, extension := range group {
			extension = strings.TrimSpace(extension)
			if extension == "" {
				return nil, fmt.Errorf("extension URI must not be empty")
			}
			if _, exists := seen[extension]; exists {
				continue
			}
			seen[extension] = struct{}{}
			dst = append(dst, extension)
		}
	}
	return dst, nil
}

func appendConditions(dst *[]yaml.Node, seen map[string]string, source string, conditions []yaml.Node) ([]string, error) {
	names := make([]string, 0, len(conditions))
	for i := range conditions {
		name, err := conditionName(&conditions[i])
		if err != nil {
			return nil, fmt.Errorf("%s condition %d: %w", source, i+1, err)
		}
		if previous, exists := seen[name]; exists {
			return nil, fmt.Errorf("condition name %q from %s collides with condition from %s", name, source, previous)
		}
		seen[name] = source
		*dst = append(*dst, conditions[i])
		names = append(names, name)
	}
	return names, nil
}

func conditionName(node *yaml.Node) (string, error) {
	if node.Kind != yaml.MappingNode {
		return "", fmt.Errorf("condition must be a mapping")
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value != "name" {
			continue
		}
		name := strings.TrimSpace(node.Content[i+1].Value)
		if name == "" {
			return "", fmt.Errorf("condition name must not be empty")
		}
		return name, nil
	}
	return "", fmt.Errorf("condition name is required")
}

func cloneMetadata(in Metadata) Metadata {
	return Metadata{
		Name:        in.Name,
		Labels:      cloneMap(in.Labels),
		Annotations: cloneMap(in.Annotations),
	}
}

func cloneMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
