package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionGreaterOrEqual(t *testing.T) {
	testCases := []struct {
		name     string
		base     string
		target   string
		expected bool
	}{
		{
			name:     "Equal",
			base:     "1.0.0",
			target:   "1.0.0",
			expected: true,
		},
		{
			name:     "Greater",
			base:     "1.0.0",
			target:   "1.0.1",
			expected: true,
		},
		{
			name:     "Lesser",
			base:     "1.0.1",
			target:   "1.0.0",
			expected: false,
		},
		{
			name:     "PreReleaseGreater",
			base:     "1.0.0-alpha",
			target:   "1.0.0-beta",
			expected: true,
		},
		{
			name:     "PreReleaseLesser",
			base:     "1.0.0-beta",
			target:   "1.0.0-alpha",
			expected: false,
		},
		{name: "EmptyBase", base: "", target: "1.0.0", expected: false},
		{name: "EmptyTarget", base: "1.0.0", target: "", expected: false},
		{name: "InvalidBase", base: "invalid", target: "1.0.0", expected: false},
		{name: "InvalidTarget", base: "1.0.0", target: "invalid", expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, VersionGreaterOrEqual(tc.base, tc.target))
		})
	}
}
