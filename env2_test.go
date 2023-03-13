package cloudy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv2(t *testing.T) {
	envSvc := NewMapEnvironment()
	envSvc.Set("ARKLOUD_V1", "root-v1")
	envSvc.Set("ARKLOUD_SERVICE_V1", "service-v1")
	envSvc.Set("ARKLOUD_V2", "root-v2")
	envSvc.Set("ARKLOUD_ONE_TWO_THREE_V1", "one-two-three-v1")

	root := NewHierarchicalEnvironment(envSvc, "ARKLOUD")
	service := root.S("service")
	ott := root.S("one", "two", "three")

	assert.Equal(t, root.Force("v1"), "root-v1")
	assert.Equal(t, service.ForceCascade("v1"), "service-v1")
	assert.Equal(t, ott.Force("V1"), "one-two-three-v1")
	assert.Equal(t, ott.ForceCascade("v2"), "root-v2")

	v2, e2 := ott.Get("v2")
	assert.NotNil(t, e2, "Key not found error should occur")
	assert.Empty(t, v2, "should be empty")
}
