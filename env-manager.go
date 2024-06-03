package cloudy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// EnvManager provides functionality to manage environment variables
type EnvManager struct {
	Prefix         string
	envDefinitions []EnvDefinition
	Sources        []EnvSource
}

// EnvDefinition represents an environment variable
type EnvDefinition struct {
	Name        	string
	Description 	string
	Keys        	[]string
	DefaultValue    string
	Value       	string
}

type EnvSource struct {
	Name	string
	Service EnvironmentService
}

var DefaultEnvManager = NewEnvManager("default")

func SetDefaultEnvManager(em *EnvManager) {
	DefaultEnvManager = em
}

func GetDefaultEnvManager() *EnvManager {
	return DefaultEnvManager
}

// NewEnvManager creates a new instance of EnvManager
func NewEnvManager(prefix string) *EnvManager {
	envMgr := &EnvManager{
		Prefix:         prefix,
		envDefinitions: make([]EnvDefinition, 0),
		Sources:        make([]EnvSource, 0),
	}

	Info(context.Background(), "EnvMgr[%s]: Created successfully", envMgr.Prefix)
	return envMgr
}


// FilteredEnvManagerByKeyPrefix returns a copy of EnvManager with filtered envDefinitions
func (em *EnvManager) FilteredEnvManagerByKeyPrefix(prefix string) *EnvManager {
	filteredEnvDefinitions := []EnvDefinition{}

	for _, def := range em.envDefinitions {
		// Filter keys by prefix
		filteredKeys := []string{}
		for _, key := range def.Keys {
			if strings.HasPrefix(key, prefix) {
				filteredKeys = append(filteredKeys, key)
			}
		}

		// If there are any keys left after filtering, add the definition to the filtered list
		if len(filteredKeys) > 0 {
			filteredEnvDefinitions = append(filteredEnvDefinitions, EnvDefinition{
				Name:        def.Name,
				Description: def.Description,
				DefaultValue:     def.DefaultValue,
				Keys:        filteredKeys,
				Value:       def.Value,
			})
		}
	}

	// Create a new EnvManager with the filtered definitions
	return &EnvManager{
		Prefix:         em.Prefix,
		envDefinitions: filteredEnvDefinitions,
		Sources:        em.Sources,
	}
}

func (em *EnvManager) validateSourceName(name string) (string, error) {
    validNamePattern := `^[a-zA-Z0-9-]+$`
    re := regexp.MustCompile(validNamePattern)

    if !re.MatchString(name) {
        return "", fmt.Errorf("EnvMgr[%s]: source name failed validation: \"%s\". Names must contain only letters, numbers, and hyphens", em.Prefix, name)
    }

    return name, nil
}

func (em *EnvManager) ValidateSourceNames(names []string) ([]string, error) {
    var validatedNames []string

    for _, name := range names {
        validatedName, err := em.validateSourceName(name)
        if err != nil {
			return nil, err
        }
		validatedNames = append(validatedNames, validatedName)
    }

	return validatedNames, nil
}

func (em *EnvManager) ValidateKey(key string) (string, error) {
    validKeyPattern := `^[A-Z0-9_]+$`
    re := regexp.MustCompile(validKeyPattern)

    if !re.MatchString(key) {
        return "", fmt.Errorf("EnvMgr[%s]: key failed validation: \"%s\". Keys must contain only uppercase letters, numbers, and underscores", em.Prefix, key)
    }

	return key, nil
}

func (em *EnvManager) ValidateKeys(keys []string) ([]string, error) {
    var validatedNames []string

    for _, name := range keys {
        validatedKey, err := em.ValidateKey(name)
        if err != nil {
			return nil, err
        }
		validatedNames = append(validatedNames, validatedKey)
    }

	return validatedNames, nil
}


func (em *EnvManager) LoadSources(sources ...string) {
	startingNumSources := len(em.Sources)

	if len(sources) == 0 {
		sources = []string{"test", "osenv", "azure-keyvault-cached"}
		Warn(context.Background(), "EnvMgr[%s]: No sources provided. Using defaults: %s", em.Prefix, sources)
	}

	sources, err := em.ValidateSourceNames(sources)
	if err != nil {
		log.Fatalf("EnvMgr[%s]: Error validating source names with error: %s", em.Prefix, err)
	}

	Info(context.Background(), "EnvMgr[%s]: Loading Sources: %s", em.Prefix, sources)

	for _, svcDriver := range sources {
		Info(context.Background(), "EnvMgr[%s]: \tadding provider [%s]", em.Prefix, svcDriver)

		// add required vars to env mgr
		providerVars, err := EnvironmentProviders.GetRequiredVars(em, svcDriver)
		if err != nil {
			log.Fatalf("EnvMgr[%s]: Could not retrieve variables required by environment: %v -> %v", em.Prefix, svcDriver, err)
		}
		if len(providerVars) > 0 {
			em.LoadVarList(providerVars)
		}

		// try to create provider, using GetVars()
		envSvcInstance, err := EnvironmentProviders.NewFromEnvMgrWith(em, svcDriver)
		if err != nil {
			log.Fatalf("EnvMgr[%s]: Could not create environment: %v -> %v", em.Prefix, svcDriver, err)
		}

		source := EnvSource{
			Name: svcDriver,
			Service: envSvcInstance,
		}
		em.Sources = append(em.Sources, source)
		Info(context.Background(), "\tEnvMgr[%s]: successfully added provider [%s]", em.Prefix, svcDriver)
	}

	NumSources := len(em.Sources)
	Info(context.Background(), "EnvMgr[%s]: Loaded [%d] new sources for a total of [%d] sources", em.Prefix, NumSources-startingNumSources, NumSources)

	// TODO: raise error if any missing during environment source loading, with complete list of provider/missing. Right now, the missing list will trigger only for a single provider
}

func (em *EnvManager) NewVarFromDef(newDef EnvDefinition) {
	Info(context.Background(), "EnvMgr[%s]: NewVarFromDef name=%s, desc=%s, default=%s", em.Prefix, newDef.Name, newDef.Description, newDef.DefaultValue)

	_, err := em.ValidateKeys(newDef.Keys)
	if err != nil {
		log.Fatalf("EnvMgr[%s]: Error validating keys with error: %s", em.Prefix, err)
	}

	for i, def := range em.envDefinitions {
		// Check if an EnvDefinition already exists in the envManager with this key
		for _, k := range def.Keys {
			if k == newDef.Keys[0] {
				if def.Name == newDef.Name {
					Warn(context.Background(), "An EnvDefinition already exists with the key \"%s\" and name \"%s\". Skipping..", newDef.Keys[0], newDef.Name)
					return
				} else {
					log.Fatalf("Attempted to re-register variable key \"%s\" under definition named \"%s\" when it was already registered under name \"%s\"", newDef.Keys[0], newDef.Name, def.Name)
				}
			}
		}

		// Key is unique now
		// Check if there is an EnvDefinition with this name. if so, add this key and return
		if def.Name == newDef.Name {
			Info(context.Background(), "EnvDefinition matching name found '%s': adding key '%s'", newDef.Name, newDef.Keys[0])
			em.envDefinitions[i].Keys = append(em.envDefinitions[i].Keys, newDef.Keys[0])
			return
		}
	}

	// If no existing EnvDefinition is found, create a new one
	em.envDefinitions = append(em.envDefinitions, newDef)
}

// NewVar creates an EnvDefinition and adds it to the EnvManager
func (em *EnvManager) NewVar(key string, name string, description string, defaultVal string) {
	Info(context.Background(), "EnvMgr[%s]: NewVar key=%s, name=%s, desc=%s, default=%s", em.Prefix, key, name, description, defaultVal)

	definition := EnvDefinition{
		Name:        name,
		Description: description,
		DefaultValue:     defaultVal,
		Keys:        []string{key},
	}

	em.NewVarFromDef(definition)
}

// Get retrieves the value of the specified environment variable from the EnvManager
// fails if an empty value is retrieved.
func (em *EnvManager) GetVar(keys ...string) string {
	for i, key := range keys {
		for _, def := range em.envDefinitions {
			for _, k := range def.Keys {
				if k == key {
					if def.Value == "" && def.DefaultValue == "" {
						log.Fatalf("EnvMgr[%s]: Environment variable %s is not set for definition name [%s]", em.Prefix, key, def.Name)
					}
					Info(context.Background(), "EnvMgr[%s]: Retrieved key \"%s\"", em.Prefix, key)
					return def.Value
				}
			}
		}

		if i < len(keys)-1 {
			Warn(context.Background(), "EnvMgr[%s]: Key \"%s\" not found, falling back to the next key", em.Prefix, key)
		} else {
			log.Fatalf("EnvMgr[%s]: None of the provided keys %v were found. Ensure they are registered with NewVar()", em.Prefix, keys)
		}
	}
	return ""
}


func (em *EnvManager) GetInt(key string) int {
	Info(context.Background(), "EnvMgr[%s]: GetInt for key [%s]", em.Prefix, key)

	vs := em.GetVar(key)
	vi, err := strconv.Atoi(vs)
	if err != nil {
		log.Fatalf("EnvMgr[%s]: GetInt conversion failed for key [%s]", em.Prefix, key)
	}
	return vi
}

func (em *EnvManager) LoadAllVars() error {
	return em.LoadVars([]string{})
}

func (em *EnvManager) LoadVarList(variableKeys []string) error {
	if len(variableKeys) == 0 {
		return errors.New("LoadVarList: cannot load empty list")
	}
	return em.LoadVars(variableKeys)
}

// retrieves the values of the specified environment variables (or all if none are specified) and stores them in the EnvDefinition
func (em *EnvManager) LoadVars(variableKeys []string) error {
	if len(em.Sources) == 0 {
		log.Fatalf("Cannot LoadVars with zero sources")
	}

	var missingVariables []EnvDefinition

	numNewLoaded := 0
	numOldLoaded := 0
	numDefault := 0

	// Determine if we need to load all variables or only specified ones
	loadAll := len(variableKeys) == 0
	variableKeySet := make(map[string]struct{}, len(variableKeys))
	if !loadAll {
		Info(context.Background(), "EnvMgr[%s]: LoadVarList() with [%d] definitions and [%d] sources", em.Prefix, len(variableKeys), len(em.Sources))

		for _, key := range variableKeys {
			variableKeySet[key] = struct{}{}
		}
	} else {
		Info(context.Background(), "EnvMgr[%s]: LoadAllVars() with [%d] definitions and [%d] sources", em.Prefix, len(em.envDefinitions), len(em.Sources))
	}

	// Iterate over all environment variable definitions
	for i := range em.envDefinitions {
		def := &em.envDefinitions[i]

		Info(context.Background(), "EnvMgr[%s]: Loading %s", em.Prefix, def.Name)
		found := false

		// Check if the variable value is already set
		if def.Value != "" {
			found = true
			Info(context.Background(), "\tEnvMgr: A value is already set for EnvDefinition named [%s]", def.Name)
			numOldLoaded++
			continue
		}

		// Iterate keys of this definition until we find a value
		for _, envSource := range em.Sources {
			for _, key := range def.Keys {
				// Check if the current definition's name is in the specified variable keys (when not loading all)
				if !loadAll {
					if _, exists := variableKeySet[key]; !exists {
						Warn(context.Background(), "EnvMgr[%s]: limited load, skipping key %s", em.Prefix, key)
						found = true
						continue
					}
				}

				Info(context.Background(), "\tEnvMgr: LoadVars() checking for key [%s] in EnvService [%s]", key, envSource.Name)
				value, err := envSource.Service.Get(key)
				if err == nil && value != "" {
					def.Value = value
					found = true
					numNewLoaded++
					Info(context.Background(), "\t\tEnvMgr: Loaded %s from EnvService [%s] with key [%s]", def.Name, envSource.Name, key)
					break
				}
			}

			if found {
				break
			}
		}

		// If no value found, use default value if available
		if def.Value == "" && def.DefaultValue != "" {
			def.Value = def.DefaultValue
			Info(context.Background(), "\tEnvMgr: Loaded %s from default value", def.Name)
			found = true
			numDefault++
			continue
		}

		// If not found and no default, add it to the list of missing variables
		if !found {
			missingVariables = append(missingVariables, *def)
		}
	}

	Info(context.Background(), "EnvMgr[%s]: LoadVars() complete with [%d] newly loaded, [%d] previously loaded and [%d] using defaults", em.Prefix, numNewLoaded, numOldLoaded, numDefault)

	// Generate missing variables list
	numMissing := len(missingVariables)
	if numMissing > 0 {
		var missingList []string
		missingList = append(missingList, fmt.Sprintf("[%d] Missing environment variables:", numMissing))
		for i, missing := range missingVariables {
			missingList = append(missingList, fmt.Sprintf("[%d] Name: %s, Key(s): %s, Description: %s", i+1, missing.Name, strings.Join(missing.Keys, ", "), missing.Description))
		}
		log.Fatalf(strings.Join(missingList, "\n"))
	}

	return nil
}