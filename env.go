package cloudy

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"unicode"
)

type Environment struct {
	m      *sync.Mutex
	prefix string
	values map[string]string
}

type SegmentedEnvironment struct {
	prefix  string
	segment string
	environ *Environment
}

func NewEnvironment(prefix string) *Environment {
	uPrefix := NormalizeEnvName(prefix)

	environ := &Environment{
		prefix: uPrefix,
		values: make(map[string]string),
	}

	return environ
}

func (environ *Environment) Root() *SegmentedEnvironment {
	return &SegmentedEnvironment{
		prefix:  environ.prefix,
		segment: "",
		environ: environ,
	}
}

func (environ *Environment) Segment(segment string) *SegmentedEnvironment {
	s := NormalizeEnvName(segment)
	newPrefix := EnvJoin(environ.prefix, s)
	return &SegmentedEnvironment{
		prefix:  newPrefix,
		segment: s,
		environ: environ,
	}
}

func (segEnv *SegmentedEnvironment) Get(name string) (string, bool) {
	return segEnv.environ.Get(name, segEnv.segment)
}

func (segEnv *SegmentedEnvironment) Default(name string, defaultValue string) (string, bool) {
	val, found := segEnv.Get(name)
	if !found {
		return defaultValue, false
	}
	return val, true
}

func (segEnv *SegmentedEnvironment) Force(name string) string {
	val, found := segEnv.Get(name)
	if !found {
		full := EnvJoin(segEnv.prefix, name)
		log.Fatalf("Required Variable not found, %v", full)
	}
	return val
}

func (segEnv *SegmentedEnvironment) GetCascade(name string, others ...string) (string, bool) {
	return segEnv.environ.GetCascade(name, segEnv.segment, others...)
}

func (segEnv *SegmentedEnvironment) ForceCascade(name string, others ...string) string {
	val, found := segEnv.GetCascade(name, others...)
	if !found {
		full := EnvJoin(segEnv.prefix, name)
		log.Fatalf("Required Variable not found, %v", full)
	}
	return val
}

func (environ *Environment) Put(name string, value string) {
	nName := NormalizeEnvName(name)
	environ.values[nName] = value
}

// Get retrieves an environment value
func (environ *Environment) Get(name string, area string) (string, bool) {
	// SKYCLOUD_AZ_TENANT_ID
	// SKYCLOUD_VMS_AZ_TENTANT_ID
	// Get("AZ_TENANT_ID", "VMS")

	raw := NormalizeEnvName(name)                       // "AZ_TENANT_ID"
	unprefixed := EnvJoin(area, name)                   // "VMS_AZ_TENANT_ID"
	fullPrefixed := EnvJoin(environ.prefix, unprefixed) // SKYCLOUD_VMS_AZ_TENTANT_ID
	rawPrefixed := EnvJoin(environ.prefix, raw)         //SKYCLOUD_AZ_TENTANT_ID

	val, foundName := environ.values[fullPrefixed]
	if foundName {
		return val, true
	}

	val, foundName = environ.values[rawPrefixed]
	if foundName {
		return val, true
	}

	val, foundName = environ.values[unprefixed]
	if foundName {
		return val, true
	}

	val, foundName = environ.values[raw]
	if foundName {
		return val, true
	}

	return "", false
}

func (environ *Environment) GetCascade(name string, area string, others ...string) (string, bool) {
	val, found := environ.Get(name, area)
	if found {
		return val, true
	}

	for _, otherName := range others {
		val, found := environ.Get(otherName, area)
		if found {
			return val, true
		}
	}

	return "", false
}

func (environ *Environment) Force(name string, area string) string {
	val, found := environ.Get(name, area)
	if !found {
		log.Fatalf("Required Variable not found, %v - %v", area, name)
	}
	return val
}

func (environ *Environment) ForceCascade(name string, area string, others ...string) string {
	val, found := environ.GetCascade(name, area, others...)
	if !found {
		log.Fatalf("Required Variable not found, %v - %v", area, name)
	}
	return val
}

// FromOSEnvironment reads the environment from the OS Environment
// variables and parsed them into the environment structure. This overwrites
// whatever is present. All the
func (environ *Environment) FromOSEnvironment() *Environment {

	// Look for all the
	for _, env := range os.Environ() {

		if env[0:1] == "\"" {
			env = env[1:]
		}
		if strings.HasSuffix(env, "\"") {
			env = env[:(len(env) - 1)]
		}
		i := strings.Index(env, "=")
		if i < 0 {
			continue
		}
		envName := NormalizeEnvName(env[0:i])
		envValue := env[i+1:]
		environ.Put(envName, envValue)
	}

	return environ
}

func (environ *Environment) FromFile(f string) *Environment {

	return environ
}

func (environ *Environment) FromJSON(vals map[string]interface{}, prev string) *Environment {
	for k, v := range vals {
		path := EnvJoin(prev, k)
		m := v.(map[string]interface{})
		if m != nil {
			return environ.FromJSON(m, path)
		} else {
			vStr := fmt.Sprintf("%v", v)
			environ.Put(path, vStr)
		}
	}
	return environ
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
