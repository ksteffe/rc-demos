package devbind

import (
	"fmt"
	"regexp"
	"strconv"
)

// The common-integrations extension allows interface.spec.version to be an
// exact MAJOR.MINOR.PATCH or one constraint: =, >, >=, <, <=, ^ or ~.
var (
	constraintPattern = regexp.MustCompile(`^(=|>=|<=|>|<|\^|~)?(\d+)\.(\d+)\.(\d+)$`)
	versionPattern    = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
)

type semver [3]int

func (v semver) compare(o semver) int {
	for i := range v {
		if v[i] != o[i] {
			if v[i] < o[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func toSemver(parts []string) semver {
	var v semver
	for i, p := range parts {
		v[i], _ = strconv.Atoi(p)
	}
	return v
}

// satisfiesVersion reports whether the concrete version satisfies constraint.
func satisfiesVersion(constraint, version string) (bool, error) {
	c := constraintPattern.FindStringSubmatch(constraint)
	if c == nil {
		return false, fmt.Errorf("version requirement %q is not a constraint this consumer understands", constraint)
	}
	v := versionPattern.FindStringSubmatch(version)
	if v == nil {
		return false, fmt.Errorf("provider version %q is not MAJOR.MINOR.PATCH", version)
	}
	want, have := toSemver(c[2:]), toSemver(v[1:])
	cmp := have.compare(want)

	switch c[1] {
	case "", "=":
		return cmp == 0, nil
	case ">":
		return cmp > 0, nil
	case ">=":
		return cmp >= 0, nil
	case "<":
		return cmp < 0, nil
	case "<=":
		return cmp <= 0, nil
	case "^":
		return cmp >= 0 && have.compare(semver{want[0] + 1, 0, 0}) < 0, nil
	default: // "~"
		return cmp >= 0 && have.compare(semver{want[0], want[1] + 1, 0}) < 0, nil
	}
}
