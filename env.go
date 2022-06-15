package cloudy

import "strings"

func EnvJoin(envParts ...string) string {
	trimmed := make([]string, len(envParts))
	for i, p := range envParts {
		trimmed[i] = strings.TrimRight(p, "_")
	}

	return strings.Join(trimmed, "_")
}
