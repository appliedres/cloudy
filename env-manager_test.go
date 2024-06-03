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

func TestNewVarFromDef(t *testing.T) {
	em := NewEnvManager("test")
	newDef := EnvDefinition{
		Name:        "test_var",
		Description: "A test variable",
		Keys:        []string{"TEST_VAR"},
		DefaultValue:     "default_value",
	}

	em.NewVarFromDef(newDef)
	assert.Len(t, em.envDefinitions, 1)
	assert.Equal(t, newDef, em.envDefinitions[0])
}

func TestNewVar(t *testing.T) {
	em := NewEnvManager("test")
	em.NewVar("TEST_KEY", "test_var", "A test variable", "default_value")

	assert.Len(t, em.envDefinitions, 1)
	assert.Equal(t, "test_var", em.envDefinitions[0].Name)
	assert.Equal(t, "TEST_KEY", em.envDefinitions[0].Keys[0])
	assert.Equal(t, "default_value", em.envDefinitions[0].DefaultValue)
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
	em.NewVar("TEST_DEFAULT_KEY", "test default key", "entry used for testing default key retrieval", testDefaultString)
	em.LoadAllVars()
	value := em.GetVar("TEST_DEFAULT_KEY")
	assert.Equal(t, testDefaultString, value)

	// test loading non-default
	em.NewVar("TEST_KEY", "test key", "entry used for testing non-default key retrieval", "")
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
	em.NewVar("TEST_DEFAULT_KEY", "test default key", "entry used for testing default key retrieval", testDefaultString)
	em.LoadAllVars()
	value := em.GetInt("TEST_DEFAULT_KEY")
	assert.Equal(t, testDefaultInt, value)

	// test loading non-default
	em.NewVar("TEST_KEY", "test key", "entry used for testing non-default key retrieval", "")
	em.LoadAllVars()
	value = em.GetInt("TEST_KEY")
	assert.Equal(t, testInt, value)
}

func TestLoadVarList(t *testing.T) {
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
	em.NewVar("UNLOADED_KEY", "unused key", "for verifying we don't load this key", "")
	em.NewVar("TEST_KEY", "test var", "A test variable", "")
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
}

func TestFilteredEnvManagerByKeyPrefix(t *testing.T) {
	testString := "test_value"

	em := NewEnvManager("test")
	em.NewVar("TEST_KEY_1", "test_var_1", "A test variable 1", "")
	em.NewVar("ANOTHER_KEY_2", "test_var_2", "A test variable 2", "")
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
	em.NewVar("TEST_KEY", "test var", "A test variable", defaultString)
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