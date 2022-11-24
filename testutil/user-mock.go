package testutil

import (
	"fmt"
	"net/http"
	"time"

	"github.com/appliedres/cloudy"
	"github.com/golang-jwt/jwt/v4"
)

func CreateTestJWT(upn string, isAdmin bool) *cloudy.UserJWT {
	jwt := &cloudy.UserJWT{
		UPN: upn,
	}
	if isAdmin {
		grp := cloudy.ForceEnv("ADMIN_GROUP", "")
		jwt.Groups = append(jwt.Groups, grp)
	}
	return jwt
}

type fakeJWT struct {
	*jwt.RegisteredClaims
	UPN    string
	Groups []string
}

func (f *fakeJWT) Valid() error {
	return nil
}

func CreateTestJWTToken(upn string, isAdmin bool) string {
	mySigningKey := []byte("AllYourBase")

	claims := &fakeJWT{
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(1516239022, 0)),
			Issuer:    "test",
		},
		UPN: upn,
	}

	if isAdmin {
		grp := cloudy.ForceEnv("ADMIN_GROUP", "")
		claims.Groups = append(claims.Groups, grp)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	fmt.Printf("%v %v", ss, err)

	return ss
}

func MockRequestWithUser(upn string, isAdmin bool) *http.Request {
	token := CreateTestJWTToken(upn, isAdmin)
	h := map[string][]string{
		"Accept-Encoding": {"gzip, deflate"},
		"Accept-Language": {"en-us"},
		"Authorization":   {"Bearer " + token},
	}

	return &http.Request{
		Method: "GET",
		Header: h,
	}
}
