package cloudy

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type UserJWT struct {
	EXP               int64                  `json:"exp"`
	IAT               int64                  `json:"iat"`
	AuthTime          int64                  `json:"auth_time"`
	JTI               string                 `json:"jti"`
	ISS               string                 `json:"iss"`
	AUD               string                 `json:"aud"`
	TYP               string                 `json:"typ"`
	AZP               string                 `json:"azp"`
	Nonce             string                 `json:"nonce"`
	SessionState      string                 `json:"session_state"`
	ACR               string                 `json:"acr"`
	AllowedOrigins    []string               `json:"allowed-origins"`
	RealmAccess       *UserJWTRealmAccess    `json:"realm_access"`
	ResourceAccess    *UserJWTResourceAccess `json:"resource_access"`
	Scope             string                 `json:"scope"`
	EmailVerified     bool                   `json:"email_verified"`
	Name              string                 `json:"name"`
	PreferredUserName string                 `json:"preferred_username"`
	GivenName         string                 `json:"given_name"`
	FamilyName        string                 `json:"family_name"`
	Email             string                 `json:"email"`
	UPN               string                 `json:"upn"`
	Groups            []string               `json:"groups"`
}

type UserJWTRealmAccess struct {
	Roles []string `json:"roles"`
}
type UserJWTResourceAccess struct {
	Account *UserJWTResourceAccessAccount `json:"account"`
}
type UserJWTResourceAccessAccount struct {
	Roles []string `json:"roles"`
}

// Valid determines if the claims are valid
func (jwt UserJWT) Valid() error {
	return nil
}

func (jwt *UserJWT) IsAuthenticated() bool {
	return !(jwt.Email == "" || jwt.Email == "None")
}

const UserAnonymous = "ANONYMOUS"

func IsAdmin(user *UserJWT) bool {
	if user.RealmAccess != nil {
		for _, role := range user.RealmAccess.Roles {
			if strings.EqualFold("DevSecOps", role) {
				return true
			}
		}
	}
	if user.Groups != nil {
		admin := GetEnv("ADMIN_GROUP", "")

		if admin != "" {
			fmt.Printf("Checking Admin %v\n", admin)

			for _, role := range user.Groups {
				if strings.EqualFold(admin, role) {
					return true
				}
			}
		} else {
			fmt.Printf("Admin Group not set\n")
		}
	}

	return false
}

func GetUserFromRequest(ctx context.Context, request *http.Request) (*UserJWT, error) {
	// fmt.Printf("Getting User Information\n")
	tokens := request.Header["Authorization"]
	if len(tokens) == 1 {
		// Info(ctx, "Found Token")
		return GetUserInfoFromToken(tokens[0]), nil
	} else if len(tokens) > 1 {
		return nil, Error(ctx, "Multiple Tokens found: %v\n", tokens)
	}

	token := request.URL.Query().Get("bearer")
	if token != "" {
		return GetUserInfoFromToken(token), nil
	}

	return nil, Error(ctx, "No Tokens found: %v\n", tokens)
}

// GetUserInfoFromToken Gets a user information from the JWT (Authorization Header)
func GetUserInfoFromToken(token string) *UserJWT {
	if token == "" || strings.EqualFold(token, "Bearer undefined") {
		return &UserJWT{
			PreferredUserName: UserAnonymous,
			Email:             "None",
		}
	}

	claims, err := ParseToken(token)
	if err != nil {
		return &UserJWT{
			PreferredUserName: "PARSING ERROR",
			Email:             "None",
		}
	}

	return claims
}

// ParseToken Parses the id token from cognito
func ParseToken(tokenstr string) (*UserJWT, error) {
	// fmt.Printf("PARSING JWT TOKEN %v\n", tokenstr)
	tokenToParse := tokenstr
	if strings.Contains(strings.ToLower(tokenstr), "bearer ") {
		tokenToParse = tokenstr[7:]
	}

	parser := new(jwt.Parser)
	var claims UserJWT

	_, _, err := parser.ParseUnverified(tokenToParse, &claims)
	if err != nil {
		// fmt.Printf("%v\n", err)
		return nil, err
	}

	return &claims, nil
}
