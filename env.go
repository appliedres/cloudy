package cloudy

import (
	"log"
	"os"
	"strings"
	"unicode"
)

func EnvJoin(envParts ...string) string {
	var trimmed []string
	for _, p := range envParts {
		v := NormalizeEnvName(p)
		v = strings.TrimRight(v, "_")
		if v != "" {
			trimmed = append(trimmed, v)
		}
	}

	return strings.Join(trimmed, "_")
}

func GetEnv(name string, prefix string) string {
	fullname := ToEnvName(name, prefix)
	return os.Getenv(fullname)
}

// ForceEnv Fails if the env is not present
func ForceEnv(name string, prefix string) string {
	val := GetEnv(name, prefix)
	if val == "" {
		log.Fatalf("No environment variable found for %v", name)
	}
	return val
}

// Normalized an env name
// MY_VAR => MY_VAR
// my-var => MY_VAR
// MyVar => MY_VAR
// myVar => MY_VAR
// my-vAR => MY_VAR
// myVAR => MY_VAR
// my var => MY_VAR
// IF there are no dashes or underscores then the first letter and virst capital in a sequence are used to have underscores
func NormalizeEnvName(v string) string {
	hasUnderscore := strings.Contains(v, "_")
	hasDash := strings.Contains(v, "-")
	hasSpace := strings.Contains(v, " ")

	if !hasDash && !hasUnderscore && !hasSpace {
		v = CamelToSeparated(v, "_")
	}

	vfixed := strings.ReplaceAll(v, "-", "_")
	vfixed = strings.ReplaceAll(vfixed, " ", "_")
	vUpper := strings.ToUpper(vfixed)
	return vUpper
}

func CamelToSeparated(v string, sep string) string {
	var newString string
	var prevBreak = false
	for i, c := range v {
		if !prevBreak && unicode.IsUpper(c) {
			prevBreak = true
			if i > 0 {
				newString += sep
			}
		} else if !unicode.IsUpper(c) {
			prevBreak = false
		}
		newString += string(c)
	}
	return newString
}

func ToEnvName(name string, prefix string) string {
	EnvJoin(prefix, name)
	fullname := name
	if prefix != "" {
		if strings.HasSuffix(prefix, "_") {
			fullname = prefix + name
		} else {
			fullname = prefix + "_" + name
		}
	}
	return NormalizeEnvName(fullname)
}
