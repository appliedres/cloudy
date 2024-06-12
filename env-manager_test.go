package cloudy

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockEnvironmentService struct {
	store map[string]string
}

func NewMockEnvironmentService() *MockEnvironmentService {
	return &MockEnvironmentService{
		store: make(map[string]string),
	}
}

func (m *MockEnvironmentService) Get(key string) (string, error) {
	if value, exists := m.store[key]; exists {
		return value, nil
	}
	return "", fmt.Errorf("key not found: %s", key)
}

func (m *MockEnvironmentService) Set(key, value string) error {
	m.store[key] = value
	return nil
}

func TestNewEnvManager(t *testing.T) {
	prefix := "test"
	em := NewEnvManager(prefix)
	assert.NotNil(t, em)
	assert.Equal(t, prefix, em.Prefix)
	assert.Empty(t, em.envDefinitions)
	assert.Empty(t, em.Sources)
}

func TestSetDefaultEnvManager(t *testing.T) {
	em := NewEnvManager("newdefault")
	SetDefaultEnvManager(em)
	assert.Equal(t, em, GetDefaultEnvManager())
}

func TestValidateSourceName(t *testing.T) {
	em := NewEnvManager("test")

	validName := "valid-source-123"
	invalidName := "Invalid@Source!"

	// Valid case
	result, err := em.validateSourceName(validName)
	assert.Nil(t, err)
	assert.Equal(t, validName, result)

	// Invalid case
	result, err = em.validateSourceName(invalidName)
	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}

func TestValidateSourceNames(t *testing.T) {
	em := NewEnvManager("test")

	invalidSources := []string{"valid-source-1", "valid-source-2", "invalid@source"}

	validatedNames, err := em.ValidateSourceNames(invalidSources)
	assert.NotNil(t, err)
	assert.Nil(t, validatedNames)

	validSources := []string{"VALID-SOURCE-1", "valid-source-2"}
	expectedValidation := []string{"valid-source-1", "valid-source-2"}

	validatedNames, err = em.ValidateSourceNames(validSources)
	assert.Nil(t, err)
	assert.Equal(t, validatedNames, expectedValidation)
}

func TestValidateKey(t *testing.T) {
	em := NewEnvManager("test")

	validKey := "VALID_KEY_NAME"
	invalidKey := "INVALID-KEY"

	// Valid case
	result, err := em.ValidateKey(validKey)
	assert.Nil(t, err)
	assert.Equal(t, validKey, result)

	// Invalid case
	result, err = em.ValidateKey(invalidKey)
	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}

func TestValidateKeys(t *testing.T) {
	em := NewEnvManager("test")

	invalidKeys := []string{"VALID_NAME123", "invalid_name!", "ANOTHER_VALID_NAME", "123_456"}

	validatedKeys, err := em.ValidateKeys(invalidKeys)
	assert.NotNil(t, err)
	assert.Nil(t, validatedKeys)

	validKeys := []string{"VALID_NAME123", "ANOTHER_VALID_NAME", "123_456"}

	validatedKeys, err = em.ValidateKeys(validKeys)
	assert.Nil(t, err)
	assert.Equal(t, validatedKeys, validKeys)
}

func TestAddDef(t *testing.T) {
	em := NewEnvManager("test")
	newDef := EnvDefinition{
		Key:          "TEST_VAR",
		Name:         "test_var",
		Description:  "A test variable",
		DefaultValue: "default_value",
	}

	em.RegisterDef(newDef)
	assert.Len(t, em.envDefinitions, 1)
	assert.Equal(t, newDef, em.envDefinitions["TEST_VAR"])
}

func TestNewVar(t *testing.T) {
	em := NewEnvManager("test")
	em.AddDef("TEST_KEY", "test_var", "A test variable", []string{}, "default_value")

	assert.Len(t, em.envDefinitions, 1)
	assert.Equal(t, "test_var", em.envDefinitions["TEST_KEY"].Name)
	assert.Equal(t, "TEST_KEY", em.envDefinitions["TEST_KEY"].Key)
	assert.Equal(t, "default_value", em.envDefinitions["TEST_KEY"].DefaultValue)
}

func TestGetVar(t *testing.T) {
	testString := "test_string"
	testDefaultString := "default_value"

	em := NewEnvManager("test")
	mockService := NewMockEnvironmentService()
	mockService.Set("TEST_KEY", testString)

	source := EnvSource{
		Name:    "mock",
		Service: mockService,
	}
	em.Sources = append(em.Sources, source)

	// test loading defaults
	em.AddDef("TEST_DEFAULT_KEY", "test default key", "entry used for testing default key retrieval", []string{}, testDefaultString)
	em.LoadAllVars()
	value := em.GetVar("TEST_DEFAULT_KEY")
	assert.Equal(t, testDefaultString, value)

	// test loading non-default
	em.AddDef("TEST_KEY", "test key", "entry used for testing non-default key retrieval", []string{}, "")
	em.LoadAllVars()
	value = em.GetVar("TEST_KEY")
	assert.Equal(t, testString, value)
}

func TestGetInt(t *testing.T) {
	testInt := 42
	testString := strconv.Itoa(testInt)

	testDefaultInt := 99
	testDefaultString := strconv.Itoa(testDefaultInt)

	em := NewEnvManager("test")
	mockService := NewMockEnvironmentService()
	mockService.Set("TEST_KEY", testString)

	source := EnvSource{
		Name:    "mock",
		Service: mockService,
	}
	em.Sources = append(em.Sources, source)

	// test loading defaults
	em.AddDef("TEST_DEFAULT_KEY", "test default key", "entry used for testing default key retrieval", []string{}, testDefaultString)
	em.LoadAllVars()
	value := em.GetInt("TEST_DEFAULT_KEY")
	assert.Equal(t, testDefaultInt, value)

	// test loading non-default
	em.AddDef("TEST_KEY", "test key", "entry used for testing non-default key retrieval", []string{}, "")
	em.LoadAllVars()
	value = em.GetInt("TEST_KEY")
	assert.Equal(t, testInt, value)
}

func TestFilteredEnvManagerByKeyPrefix(t *testing.T) {
	testString := "test_value"

	em := NewEnvManager("test")
	em.AddDef("TEST_KEY_1", "test_var_1", "A test variable 1", []string{}, "")
	em.AddDef("ANOTHER_KEY_2", "test_var_2", "A test variable 2", []string{}, "")
	mockService := NewMockEnvironmentService()
	mockService.Set("TEST_KEY_1", testString)
	em.Sources = []EnvSource{
		{
			Name:    "test",
			Service: mockService,
		},
	}

	filteredEm := em.FilteredEnvManagerByKeyPrefix("TEST")
	assert.Len(t, filteredEm.envDefinitions, 1)

	filteredEm.LoadAllVars()

	value := filteredEm.GetVar("TEST_KEY_1")
	assert.Equal(t, value, testString)
}

func TestGetDefault(t *testing.T) {
	defaultString := "default_val"

	em := NewEnvManager("test")
	em.AddDef("TEST_KEY", "test var", "A test variable", []string{}, defaultString)
	mockService := NewMockEnvironmentService()
	em.Sources = []EnvSource{
		{
			Name:    "test",
			Service: mockService,
		},
	}

	err := em.LoadAllVars()
	assert.NoError(t, err)

	value := em.GetVar("TEST_KEY")
	assert.Equal(t, defaultString, value)
}

func TestLoadEmptyVarList(t *testing.T) {
	em := NewEnvManager("test")
	em.Sources = []EnvSource{
		{
			Name:    "test",
			Service: NewMockEnvironmentService(),
		},
	}

	err := em.LoadVarList([]string{})
	assert.Error(t, err, "LoadVarList: cannot load empty list")
}

func TestLoadVars(t *testing.T) {
	testString := "test_value"

	em := NewEnvManager("test")
	em.AddDef("UNLOADED_KEY", "unused key", "for verifying we don't load this key", []string{}, "")
	em.AddDef("TEST_KEY", "test var", "A test variable", []string{}, "")
	mockService := NewMockEnvironmentService()
	mockService.Set("TEST_KEY", testString)
	em.Sources = []EnvSource{
		{
			Name:    "test",
			Service: mockService,
		},
	}

	err := em.LoadVars([]string{"TEST_KEY"})
	assert.NoError(t, err)

	value := em.GetVar("TEST_KEY")
	assert.Equal(t, testString, value)

	value, err = em.getVar("UNLOADED_KEY")
	assert.Error(t, err)
	assert.Equal(t, "", value)
}

func TestLoadVars_AllVariables(t *testing.T) {
	envService := NewMockEnvironmentService()
	envService.Set("VAR1", "value1")
	envService.Set("VAR2", "value2")

	em := NewEnvManager("test")
	em.AddDef("VAR1", "Variable 1", "A test variable 1", []string{}, "")
	em.AddDef("VAR2", "Variable 2", "A test variable 2", []string{}, "")

	em.Sources = []EnvSource{
		{Name: "mockService", Service: envService},
	}

	err := em.LoadAllVars()
	assert.NoError(t, err)
	assert.Equal(t, "value1", em.GetVar("VAR1"))
	assert.Equal(t, "value2", em.GetVar("VAR2"))
}

func TestLoadVars_SpecificVariables(t *testing.T) {
	envService := NewMockEnvironmentService()
	envService.Set("VAR1", "value1")
	envService.Set("VAR2", "value2")

	em := NewEnvManager("test")
	em.AddDef("VAR1", "Variable 1", "A test variable 1", []string{}, "")
	em.AddDef("VAR2", "Variable 2", "A test variable 2", []string{}, "")

	em.Sources = []EnvSource{
		{Name: "mockService", Service: envService},
	}

	err := em.LoadVars([]string{"VAR1"})
	assert.NoError(t, err)
	assert.Equal(t, "value1", em.GetVar("VAR1"))

	val, err := em.getVar("VAR2")
	assert.Error(t, err)
	assert.Equal(t, "", val)
}

func TestLoadVars_DefaultValues(t *testing.T) {
	envService := NewMockEnvironmentService()

	em := NewEnvManager("test")
	em.AddDef("VAR1", "Variable 1", "A test variable 1", []string{}, "default1")
	em.AddDef("VAR2", "Variable 2", "A test variable 2", []string{}, "default2")

	em.Sources = []EnvSource{
		{Name: "mockService", Service: envService},
	}

	err := em.LoadVars(nil)
	assert.NoError(t, err)
	assert.Equal(t, "default1", em.GetVar("VAR1"))
	assert.Equal(t, "default2", em.GetVar("VAR2"))
}

func TestLoadVars_OverridesDefaultValues(t *testing.T) {
	envService := NewMockEnvironmentService()
	envService.Set("VAR1", "value1")

	em := NewEnvManager("test")
	em.AddDef("VAR1", "Variable 1", "A test variable 1", []string{}, "default1")
	em.AddDef("VAR2", "Variable 2", "A test variable 2", []string{}, "default2")

	em.Sources = []EnvSource{
		{Name: "mockService", Service: envService},
	}

	err := em.LoadVars(nil)
	assert.NoError(t, err)
	assert.Equal(t, "value1", em.GetVar("VAR1"))
	assert.Equal(t, "default2", em.GetVar("VAR2"))
}