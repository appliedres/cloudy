package cloudy

import (
	"context"
	"log"
	"strconv"
)

var EnvironmentProviders = NewProviderRegistry[EnvironmentService]()

type Environment struct {
	EnvSvc      EnvironmentService
	Credentials *CredentialManager
}

func NewEnvironment(envSvc EnvironmentService) *Environment {
	return &Environment{
		EnvSvc:      envSvc,
		Credentials: NewCredentialManager(),
	}
}

// func (env *Environment) LoadCredentials(prefix string) *CredentialManager {
// 	Info(context.Background(), "LoadCredentials prefix:%s", prefix)

// 	credEnv := env.Segment(prefix)
// 	rtn := NewCredentialManager()
// 	for source, loader := range CredentialSources {
// 		c := loader.ReadFromEnvMgr(credEnv)
// 		rtn.credentials[source] = c
// 	}
// 	return rtn
// }

func (env *Environment) GetCredential(sourceName string) interface{} {
	Info(context.Background(), "GetCredential sourceName:%s", sourceName)

	return env.Credentials.Get(sourceName)
}

func (env *Environment) Get(name string) string {
	Info(context.Background(), "Get name:%s", name)

	v, err := env.EnvSvc.Get(name)
	if err == nil && v != "" {
		return v
	}
	return ""
}

func (env *Environment) GetInt(name string) (int, bool) {
	Info(context.Background(), "GetInt name:%s", name)

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
	Info(context.Background(), "Environment Force %s", name)

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
	log.Fatalf("env2 Force Required Variable not found, %v (%s)", full, name)

	return ""
}

func (env *Environment) Default(name string, defaultValue string) string {
	Info(context.Background(), "Default name:%s", name)

	val := env.Get(name)
	if val == "" {
		return defaultValue
	}
	return val
}

func (env *Environment) Segment(name ...string) *Environment {
	var rtn *Environment
	svc := env.EnvSvc.(*HierarchicalEnvironment)
	if svc != nil {
		rtn = NewEnvironment(svc.S(name...))
	} else {
		rtn = NewEnvironment(NewHierarchicalEnvironment(svc, name...))
	}
	rtn.Credentials = env.Credentials
	return rtn
}

func (env *Environment) SegmentWithCreds(creds *CredentialManager, name ...string) *Environment {
	var rtn *Environment
	svc := env.EnvSvc.(*HierarchicalEnvironment)
	if svc != nil {
		rtn = NewEnvironment(svc.S(name...))
	} else {
		rtn = NewEnvironment(NewHierarchicalEnvironment(svc, name...))
	}
	rtn.Credentials = creds
	return rtn
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
	Info(context.Background(), "S ?? segments:%s", segments)

	h := &HierarchicalEnvironment{
		environ: env,
		prefix:  EnvJoin(segments...),
	}

	return h
}

func (segEnv *HierarchicalEnvironment) S(name ...string) *HierarchicalEnvironment {
	Info(context.Background(), "S ?? name:%s", name)
	
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
	Info(context.Background(), "H GetNoCascade %s", name)

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
	Info(context.Background(), "H Default name:%s, defualt:%s", name, defaultValue)

	val, found := segEnv.Get(name)
	if found != nil || val == "" {
		return defaultValue, false
	}
	return val, true
}

func (segEnv *HierarchicalEnvironment) ForceNoCascadee(name string) string {
	Info(context.Background(), "H ForceNoCascadee %s", name)

	val, found := segEnv.GetNoCascade(name)
	if !found {
		full := EnvJoin(segEnv.prefix, name)
		log.Fatalf("ForceNoCascadee Required Variable not found, %v", full)
	}
	return val
}

func (segEnv *HierarchicalEnvironment) Get(name string) (string, error) {
	Info(context.Background(), "HierarchicalEnvironment Get %s", name)

	raw := NormalizeEnvName(name)
	full := EnvJoin(segEnv.prefix, raw)

	v, err := segEnv.environ.Get(full)
	if err == nil {
		return v, nil
	}

	return "", ErrKeyNotFound
}

func (segEnv *HierarchicalEnvironment) Force(name string) string {
	Info(context.Background(), "HierarchicalEnvironment Force %s", name)

	val, err := segEnv.Get(name)
	if err != nil {
		full := EnvJoin(segEnv.prefix, name)
		log.Fatalf("HierarchicalEnvironment Force Required Variable not found, %v", full)
	}
	return val
}

func (segEnv *HierarchicalEnvironment) GetCascade(name string) (string, error) {
	Info(context.Background(), "HierarchicalEnvironment Get %s", name)

	raw := NormalizeEnvName(name)
	full := EnvJoin(segEnv.prefix, raw)

	v, err := segEnv.environ.Get(full)
	if v != "" {
		return v, err
	}

	if segEnv.parent != nil {
		return segEnv.parent.GetCascade(name)
	}

	return "", ErrKeyNotFound
}

func (segEnv *HierarchicalEnvironment) ForceCascade(name string) string {
	Info(context.Background(), "HierarchicalEnvironment Force %s", name)

	val, err := segEnv.GetCascade(name)
	if err != nil {
		full := EnvJoin(segEnv.prefix, name)
		log.Fatalf("Error getting variable, %v, %v", full, err)
	}
	if val == "" {
		full := EnvJoin(segEnv.prefix, name)
		log.Fatalf("HierarchicalEnvironment ForceCascade Required Variable not found, %v", full)

	}
	return val
}

type CredentialManager struct {
	credentials map[string]interface{}
}

func NewCredentialManager() *CredentialManager {
	return &CredentialManager{
		credentials: make(map[string]interface{}),
	}
}

func (creds *CredentialManager) Get(sourceName string) interface{} {
	return creds.credentials[sourceName]
}

func (creds *CredentialManager) Put(sourceName string, obj interface{}) {
	creds.credentials[sourceName] = obj
}
