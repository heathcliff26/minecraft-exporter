package utils

import (
	"log/slog"

	goversion "github.com/hashicorp/go-version"
)

const (
	VERSION_1_15   = "1.15.0"
	VERSION_1_20_3 = "1.20.3"
	VERSION_26     = "26.0.0"
)

// Compares the choosen base version against the target version.
// Returns true if the target is greater or equal.
// Always returns false if the version string is not a proper semver version.
func VersionGreaterOrEqual(base, target string) bool {
	if base == "" || target == "" {
		slog.Warn("Could not compare version strings, either the base or the target was empty", slog.String("base", base), slog.String("target", target))
		return false
	}

	vBase, err := goversion.NewSemver(base)
	if err != nil {
		slog.Warn("Failed to convert base version to semver", "error", err)
		return false
	}

	vTarget, err := goversion.NewSemver(target)
	if err != nil {
		slog.Warn("Failed to convert target version to semver", "error", err)
		return false
	}

	return vBase.LessThanOrEqual(vTarget)
}
