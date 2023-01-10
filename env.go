package cloudy

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode"
)

type EnvironmentService interface {
	Get(name string) (string, error)
}

var DefaultEnvironment = NewEnvironment(NewOsEnvironmentService())

func SetDefaultEnvironment(env *Environment) {
	DefaultEnvironment = env
}

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

func RemoveEnvPrefix(prefix string, name string) string {
	if strings.HasPrefix(name, prefix) && len(name) > len(prefix) {
		return name[len(prefix)+1:]
	}
	return name
}

func AddEnvPrefix(prefix string, name string) string {
	if strings.HasPrefix(name, prefix) && len(name) > len(prefix) {
		return name
	}
	return EnvJoin(prefix, name)
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

func LoadEnv(file string) error {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	all := string(data)
	lines := strings.Split(all, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Comment
			continue
		}
		if len(trimmed) == 0 {
			continue
		}

		index := strings.Index(trimmed, "=")
		if index > 0 {
			k := trimmed[0:index]
			v := trimmed[index+1:]

			name := NormalizeEnvName(k)
			os.Setenv(name, v)
		}

	}
	return nil
}

func CreateCompleteEnvironment(envVar string, PrefixVar string) *Environment {
	ctx := context.Background()

	// create a simple env first
	Info(ctx, "CreateCompleteEnvironment: Simple First")
	tempEnv := NewEnvironment(NewTieredEnvironment(NewTestFileEnvironmentService(), NewOsEnvironmentService()))
	envServiceList := tempEnv.Default(envVar, "test|osenv")
	prefix := tempEnv.Get(PrefixVar)

	// Split and iterate
	Info(ctx, "CreateCompleteEnvironment: Loading: %s", envServiceList)
	envServiceDrivers := strings.Split(envServiceList, "|")

	// Create the overall environment
	envServices := make([]EnvironmentService, len(envServiceDrivers))
	for i, svcDriver := range envServiceDrivers {
		envSvcInstance, err := EnvironmentProviders.NewFromEnvWith(tempEnv, svcDriver)
		if err != nil {
			log.Fatalf("Could not create environment: %v -> %v", svcDriver, err)
		}
		envServices[i] = envSvcInstance
	}

	return NewEnvironment(NewHierarchicalEnvironment(NewTieredEnvironment(envServices...), prefix))
}
