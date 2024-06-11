package cloudy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// EnvManager provides functionality to manage environment variables
type EnvManager struct {
	Prefix         string
	envDefinitions map[string]EnvDefinition
	Sources        []EnvSource
}

// EnvDefinition represents an environment variable
// Priority of retrieval AKA loading:
// 1 - check if definition already has a value
// 2 - attempt to retrieve from primary key from any source
// 3 - attempt to retrieve value from fallback keys in order
// 4 - use default value if no value found
// 5 - No values retrieved and no default, then add to missing list
type EnvDefinition struct {
	Key          string
	Name         string
	Description  string
	Value        string
	FallbackKeys []string  // TODO: implement fallback keys
	DefaultValue string
}

type EnvSource struct {
	Name    string
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
		envDefinitions: make(map[string]EnvDefinition),
		Sources:        make([]EnvSource, 0),
	}

	Info(context.Background(), "EnvMgr[%s]: Created successfully", envMgr.Prefix)
	return envMgr
}

// FilteredEnvManagerByKeyPrefix returns a copy of EnvManager with filtered envDefinitions
func (em *EnvManager) FilteredEnvManagerByKeyPrefix(prefix string) *EnvManager {
	filteredEnvDefinitions := make(map[string]EnvDefinition)

	for key, def := range em.envDefinitions {
		if strings.HasPrefix(def.Key, prefix) {
			filteredEnvDefinitions[key] = def
		}
	}

	// Create a new EnvManager with the filtered definitions
	return &EnvManager{
		Prefix:         em.Prefix,
		envDefinitions: filteredEnvDefinitions,
		Sources:        em.Sources,
	}
}

// source names are always lowercase
func (em *EnvManager) validateSourceName(name string) (string, error) {
	name = strings.ToLower(name)

	validNamePattern := `^[a-z0-9-]+$`
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

// Loads the given environment sources into the EnvManager, ensuring that each source's requisite definitions have loaded values.
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
			Name:    svcDriver,
			Service: envSvcInstance,
		}
		em.Sources = append(em.Sources, source)
		Info(context.Background(), "\tEnvMgr[%s]: successfully added provider [%s]", em.Prefix, svcDriver)
	}

	NumSources := len(em.Sources)
	Info(context.Background(), "EnvMgr[%s]: Loaded [%d] new sources for a total of [%d] sources", em.Prefix, NumSources-startingNumSources, NumSources)

	// TODO: raise error if any missing during environment source loading, with complete list of provider/missing. Right now, the missing list will trigger only for a single provider
}

// Adds a given EnvDefinition to the EnvManager, ensuring that it is a unique entry.
func (em *EnvManager) RegisterDef(newDef EnvDefinition) {
	Info(context.Background(), "EnvMgr[%s]: AddDef name=%s, desc=%s, default=%s", em.Prefix, newDef.Name, newDef.Description, newDef.DefaultValue)

	_, err := em.ValidateKey(newDef.Key)
	if err != nil {
		log.Fatalf("EnvMgr[%s]: Error validating key with error: %s", em.Prefix, err)
	}

	// Check if an EnvDefinition already exists in the envManager with this key
	if existingDef, exists := em.envDefinitions[newDef.Key]; exists {
		if reflect.DeepEqual(existingDef, newDef) {
			Warn(context.Background(), "Attempted to re-register variable key \"%s\" under definition named \"%s\" when it was already registered under name \"%s\". Since these registrations are equivalent, we are ignoring this error.", newDef.Key, newDef.Name, existingDef.Name)
		} else {
			log.Fatalf("Attempted to re-register variable key \"%s\" under definition named \"%s\" when it was already registered under name \"%s\". These registrations do not match exactly, so this fails.", newDef.Key, newDef.Name, existingDef.Name)
		}
	}

	// If no existing EnvDefinition is found, add the new one to the map
	em.envDefinitions[newDef.Key] = newDef
}

// Ceates an EnvDefinition and passes it to RegisterDef
func (em *EnvManager) AddDef(key string, name string, description string, fallbackKeys []string, defaultVal string) {
	Info(context.Background(), "EnvMgr[%s]: RegisterDef key=%s, name=%s, desc=%s, default=%s", em.Prefix, key, name, description, defaultVal)

	definition := EnvDefinition{
		Key:          key,
		Name:         name,
		Description:  description,
		FallbackKeys: fallbackKeys,
		DefaultValue: defaultVal,
	}

	em.RegisterDef(definition)
}

// Get retrieves the value of the specified environment variable from the EnvManager
// fails if an empty value is retrieved.
func (em *EnvManager) GetVar(keys ...string) string {
	Info(context.Background(), "EnvMgr[%s]: GetVar keys=%s", em.Prefix, keys)

	for i, key := range keys {
		if def, exists := em.envDefinitions[key]; exists {
			if def.Value == "" && def.DefaultValue == "" {
				log.Fatalf("EnvMgr[%s]: Environment variable %s is not set for definition name [%s]", em.Prefix, key, def.Name)
			}
			Info(context.Background(), "EnvMgr[%s]: Retrieved key \"%s\"", em.Prefix, key)
			return def.Value
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

	start := time.Now() // Start measuring execution time

	// Determine if we need to load all variables or only specified ones
	loadAll := len(variableKeys) == 0

	var missingVariables []EnvDefinition

	// numNewLoaded := 0
	// numFallbacksLoaded := 0
	// numPreviouslyLoaded := 0
	// numDefault := 0

	visited := make(map[string]bool) // track visited definitions to detect circular references

	if loadAll {
		Info(context.Background(), "EnvMgr[%s]: LoadAllVars() with [%d] definitions and [%d] sources", em.Prefix, len(em.envDefinitions), len(em.Sources))
	} else {
		Info(context.Background(), "EnvMgr[%s]: LoadVarList() with [%d] definitions and [%d] sources", em.Prefix, len(variableKeys), len(em.Sources))
	}

	// Iterate over all environment variable definitions
    for key, def := range em.envDefinitions {
        // If loading all or key is in variableKeys
        if loadAll || contains(variableKeys, key) {		// Pop the last item from the stack
            if def.Value != "" {
				Warn(context.Background(), "key [%s] already loaded, skipping", key)
			}

			err := em.loadVar(key, visited)
			if err != nil {
				// Step 4 - If not retrieved and no default, add it to the list of missing variables
				Warn(context.Background(), "\tEnvMgr: Adding %s to missing list", key)
				missingVariables = append(missingVariables, def)
			}

			Info(context.Background(), "key [%s] loaded", key)
		}
	}

	// Info(context.Background(), "EnvMgr[%s]: LoadVars() complete with [%d] newly loaded, [%d] previously loaded and [%d] using defaults", em.Prefix, numNewLoaded, numPreviouslyLoaded, numDefault)
	Info(context.Background(), "EnvMgr[%s]: LoadVars() complete", em.Prefix)

	// Print execution time
    end := time.Now()
	seconds := float64(end.Sub(start)) / float64(time.Second)
	roundedSeconds := fmt.Sprintf("%.2f", seconds)
	Info(context.Background(), "LoadVars() Execution time: %s seconds", roundedSeconds)

	// Generate missing variables list
	numMissing := len(missingVariables)
	if numMissing > 0 {
		var missingList []string
		missingList = append(missingList, fmt.Sprintf("[%d] Missing environment variables:", numMissing))
		for i, missing := range missingVariables {
			missingList = append(missingList, fmt.Sprintf("[%d] Name: %s, Key(s): %s, Description: %s", i+1, missing.Name, missing.Key, missing.Description))
		}
		log.Fatalf(strings.Join(missingList, "\n"))
	}

	return nil
}

// loads a single variable and will recursively load FallbackKeys with protection against circular references
func (em *EnvManager) loadVar(key string, visited map[string]bool) error {
    def, ok := em.envDefinitions[key]
    if !ok {
		log.Fatalf("definition not found for key [%s]", key)
    }

	if visited[key] {
		log.Fatalf("circular reference found when loading fallback key [%s]", key)
	}
	visited[key] = true

	Info(context.Background(), "EnvMgr[%s]: Loading %s", em.Prefix, key)

    // Step 1 - Check if the definition already has a value
    if def.Value != "" {
		Info(context.Background(), "\tEnvMgr: A value is already set for EnvDefinition named [%s]", def.Name)
        return nil
    }

	// Step 2 - Attempt to retrieve primary key from the sources
	for _, envSource := range em.Sources {
		Info(context.Background(), "\tEnvMgr: LoadVars() attempting to retrieve key [%s] from EnvService [%s]", key, envSource.Name)
		value, err := envSource.Service.Get(key)
		if err == nil && value != "" {
			def.Value = value
			em.envDefinitions[key] = def
			// TODO numNewLoaded++
			Info(context.Background(), "\t\tEnvMgr: Loaded %s from EnvService [%s]", key, envSource.Name)
			return nil
		}
	}

	// Step 3 - recursively retrieve from fallback keys if no value could be retrieved from primary key
	for _, fallbackKey := range def.FallbackKeys {
		fallbackDef, fallbackDefRegistered := em.envDefinitions[fallbackKey]

		if !fallbackDefRegistered {
			Warn(context.Background(), "Key [%s] is using fallback key [%s] that has not been registered. Skipping...", key, fallbackKey)
			continue
		}

		// if fallback def has a value, use that for our current def
		if fallbackDef.Value != "" {
			def.Value = fallbackDef.Value
			em.envDefinitions[key] = def
			// TODO numFallbacksLoaded++
			Info(context.Background(), "\t\tEnvMgr: Loaded key [%s] from fallback key [%s]", key, fallbackKey)
			return nil
		}

		// fallback definition does not have a value, load it
		err := em.loadVar(fallbackDef.Key, visited)
		if err != nil {
			Warn(context.Background(), "Error when retrieving fallback key [%s] that has not been registered. Skipping...", fallbackKey)
		}

		fallbackDef = em.envDefinitions[fallbackKey] // update the def now that it's loaded
		if fallbackDef.Value == "" {
			Warn(context.Background(), "Key [%s] used fallback key [%s] but it has no value. Skipping...", key, fallbackKey)
			continue
		}

		def.Value = fallbackDef.Value
		em.envDefinitions[key] = def
		Info(context.Background(), "Key [%s] found value from loaded fallback key [%s].", key, fallbackKey)
		return nil
	}

	// Step 4 - If no value found, use default value if available
	if def.DefaultValue != "" {
		def.Value = def.DefaultValue
		em.envDefinitions[key] = def
		Info(context.Background(), "\tEnvMgr: Loaded %s from default value", key)
		// TODO numDefault++
		return nil
	}

	// 5 - No values retrieved and no default
	Warn(context.Background(), "key [%s] could not be loaded", key)
	return fmt.Errorf("key [%s] could not be loaded", key)
}

func contains(slice []string, str string) bool {
    for _, s := range slice {
        if s == str {
            return true
        }
    }
    return false
}