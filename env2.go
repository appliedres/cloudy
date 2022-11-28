package cloudy

import (
	"log"
	"strconv"
)

var EnvironmentProviders = NewProviderRegistry[EnvironmentService]()

type Environment struct {
	EnvSvc EnvironmentService
}

func NewEnvironment(envSvc EnvironmentService) *Environment {
	return &Environment{
		EnvSvc: envSvc,
	}
}

func (env *Environment) Get(name string) string {
	v, err := env.EnvSvc.Get(name)
	if err == nil && v != "" {
		return v
	}
	return ""
}

func (env *Environment) GetInt(name string) (int, bool) {
	v, err := env.EnvSvc.Get(name)
	if err == nil && v != "" {
		vi, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}

		return vi, true
	}
	return 0, false
}

func (env *Environment) Force(name ...string) string {
	if len(name) == 0 {
		log.Fatalf("no names passed")
	}

	//looks for multlpe key values
	for _, n := range name {
		v, err := env.EnvSvc.Get(n)
		if err == nil && v != "" {
			return v
		}
	}

	full := NormalizeEnvName(name[0])
	log.Fatalf("Required Variable not found, %v", full)

	return ""
}

func (env *Environment) Default(name string, defaultValue string) string {
	val := env.Get(name)
	if val == "" {
		return defaultValue
	}
	return val
}

func (env *Environment) Segment(name ...string) *Environment {
	svc := env.EnvSvc.(*HierarchicalEnvironment)
	if svc != nil {
		return NewEnvironment(svc.S(name...))
	}
	return NewEnvironment(NewHierarchicalEnvironment(svc, name...))
}

// HierarchicalEnvironment understands that the envrionment can be split into heirarchical
// segments (a tree) that can be used to provide overrides. For isntance, ARKLOUD_V1 and ARKLOUD_SERVICE1_V1
// are both possible keys with segements being ARKLOUD and ARKLOUD_SERVICE1. So a call to
// GetCascade("V1")
type HierarchicalEnvironment struct {
	prefix  string
	environ EnvironmentService
	parent  *HierarchicalEnvironment
}

func NewHierarchicalEnvironment(env EnvironmentService, segments ...string) *HierarchicalEnvironment {
	h := &HierarchicalEnvironment{
		environ: env,
		prefix:  EnvJoin(segments...),
	}

	return h
}

func (segEnv *HierarchicalEnvironment) S(name ...string) *HierarchicalEnvironment {

	nameReal := segEnv.prefix
	last := segEnv
	for _, namepart := range name {
		nameReal = EnvJoin(nameReal, namepart)
		h := &HierarchicalEnvironment{
			environ: segEnv.environ,
			prefix:  nameReal,
			parent:  last,
		}
		last = h
	}

	return last
}

func (segEnv *HierarchicalEnvironment) GetNoCascade(name string) (string, bool) {
	raw := NormalizeEnvName(name)
	full := EnvJoin(segEnv.prefix, raw)

	v, err := segEnv.environ.Get(full)
	if err == nil {
		return v, true
	}

	// v, err = segEnv.environ.Get(raw)
	// if err == nil {
	// 	return v, true
	// }

	return v, false
}

func (segEnv *HierarchicalEnvironment) Default(name string, defaultValue string) (string, bool) {
	val, found := segEnv.Get(name)
	if found != nil || val == "" {
		return defaultValue, false
	}
	return val, true
}

func (segEnv *HierarchicalEnvironment) ForceNoCascadee(name string) string {
	val, found := segEnv.GetNoCascade(name)
	if !found {
		full := EnvJoin(segEnv.prefix, name)
		log.Fatalf("Required Variable not found, %v", full)
	}
	return val
}

func (segEnv *HierarchicalEnvironment) Get(name string) (string, error) {
	val, found := segEnv.GetNoCascade(name)
	if found {
		return val, nil
	}

	if segEnv.parent != nil {
		return segEnv.parent.Get(name)
	}

	return "", ErrKeyNotFound
}

func (segEnv *HierarchicalEnvironment) Force(name string) string {
	val, err := segEnv.Get(name)
	if err != nil {
		full := EnvJoin(segEnv.prefix, name)
		log.Fatalf("Required Variable not found, %v", full)
	}
	return val
}
