package cloudy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var fakeJwt = "eyJhbGciOiJIUzI1NiJ9.eyJSb2xlIjoiQWRtaW4iLCJ1cG4iOiJUZXN0LlVzZXJAdXBuLmNvbSIsIklzc3VlciI6Iklzc3VlciIsInByZWZlcnJlZF91c2VybmFtZSI6IlVzZXJOYW1lIiwiZXhwIjoxNjU0MDE0MDIxLCJpYXQiOjE2NTQwMTQwMjEsImVtYWlsIjoiVGVzdC5Vc2VyQGVtYWlsLmNvbSJ9.zUKGWpoo6xM4gzniTUwsV3wSK-z7xBVNoKvEWNgoupw"

func TestJWT(t *testing.T) {
	jwt, err := ParseToken(fakeJwt)

	assert.Nil(t, err)
	assert.True(t, jwt.IsAuthenticated())
	assert.Equal(t, jwt.Email, "Test.User@email.com")
	assert.Equal(t, jwt.UPN, "Test.User@upn.com")
}
