package cloudy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestEnv(t *testing.T) {

// 	os.Setenv("UNPREFIXED", "unprefixed-value")
// 	os.Setenv("PREFIXED_V1", "V1")
// 	os.Setenv("PREFIXED_V2_MORE_SPACE", "v2MoreSpace")
// 	os.Setenv("PREFIXED_AREA_V3", "V3")
// 	os.Setenv("PREFIXED_UG", "V1")
// 	os.Setenv("PREFIXED_UG_TENTANT", "V1")
// 	os.Setenv("PREFIXED_UG_USER", "V1")
// 	os.Setenv("PREFIXED_UG_GROUP", "V1")

// 	// env := NewEnvironment().FromOSEnvironment()
// 	// mine := env.S("PREFIXED").S("UG").S("USER")
// 	// mine := env.Segments("PREFIXED", "UG", "USER")

// 	env := NewEnvironment()
// 	env.FromOSEnvironment()

// 	v1, f1 := env.Root().Get("V1")
// 	assert.True(t, f1)
// 	assert.Equal(t, v1, "V1")

// 	v2, f2 := env.Root().Get("v2MoreSpace")
// 	assert.True(t, f2)
// 	assert.Equal(t, v2, "v2MoreSpace")

// 	senv := env.Segment("AREA")
// 	v3, f3 := senv.Get("V3")
// 	assert.True(t, f3)
// 	assert.Equal(t, v3, "V3")

// 	v1, f1 = senv.Get("V1")
// 	assert.True(t, f1)
// 	assert.Equal(t, v1, "V1")

// 	v1c, f1c := senv.GetCascade("MISSING", "V1")
// 	assert.True(t, f1c)
// 	assert.Equal(t, v1c, "V1")

// 	v1d, f1d := senv.Default("MISSING", "DEFAULT")
// 	assert.False(t, f1d)
// 	assert.Equal(t, v1d, "DEFAULT")

// 	v2d, f2d := senv.Default("V1", "DEFAULT")
// 	assert.True(t, f2d)
// 	assert.Equal(t, v2d, "V1")

// 	v1 = env.Root().Force("V1")
// 	assert.Equal(t, v1, "V1")

// 	// env.Root().Force("MISSING")
// 	// senv.Force("MISSING")

// 	assert.Equal(t, "AZ_TENANT_ID", NormalizeEnvName("AZ_TENANT_ID"))
// 	assert.Equal(t, "AZ_TENANT_ID", NormalizeEnvName("azTenantID"))
// 	assert.Equal(t, "AZ_TENANT_ID", NormalizeEnvName("AZ_Tenant-ID"))
// 	assert.Equal(t, "AZ_TENANT_ID", NormalizeEnvName("AzTenantId"))
// 	assert.Equal(t, "AZ_TENANT_ID", NormalizeEnvName("Az-Tenant-Id"))

// 	assert.Equal(t, "AZ_TENANT_ID", EnvJoin("az", "tenant", "id"))
// 	assert.Equal(t, "AZ_TENANT_ID", EnvJoin("azTenant", "id"))
// 	assert.Equal(t, "AZ_TENANT_ID", EnvJoin("azTenantID", ""))
// 	assert.Equal(t, "AZ_TENANT_ID", EnvJoin("", "azTenantID", ""))
// }

func TestExample(t *testing.T) {
	m := make(map[string]string)
	m["SKYCLOUD_UG_USER_DRIVER"] = "1"
	m["SKYCLOUD_UG_V2"] = "2"
	m["SKYCLOUD_V3"] = "3"

	maps := NewMaps(m)

	user := maps.S("SKYCLOUD", "UG", "USER")
	assert.Equal(t, user.Get("DRIVER"), "1")
	assert.Equal(t, user.Get("V2"), "2")
	assert.Equal(t, user.Get("V3"), "3")

	user = maps.S("SKYCLOUD").S("UG").S("USER")
	assert.Equal(t, user.Get("DRIVER"), "1")
	assert.Equal(t, user.Get("V2"), "2")
	assert.Equal(t, user.Get("V3"), "3")

}

type Maps struct {
	fullSegment string
	segment     string
	parts       []string
	parent      *Maps
	root        map[string]string
}

func NewMaps(root map[string]string) *Maps {
	return &Maps{
		root: root,
	}
}

func (maps *Maps) S(segments ...string) *Maps {
	var last = maps
	for _, seg := range segments {
		last = last.seg(seg)
	}
	return last
}

func (maps *Maps) seg(s string) *Maps {
	allParts := append(maps.parts, s)
	full := EnvJoin(allParts...)
	return &Maps{
		fullSegment: full,
		segment:     s,
		parts:       allParts,
		parent:      maps,
		root:        maps.root,
	}
}

func (maps *Maps) Get(key string) string {
	name := EnvJoin(maps.fullSegment, key)

	v, found := maps.root[name]
	if found {
		return v
	}

	if maps.parent != nil {
		return maps.parent.Get(key)
	}

	return ""
}
