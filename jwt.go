package cloudy

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserJWT struct {
	EXP      int64  `json:"exp"`
	IAT      int64  `json:"iat"`
	AuthTime int64  `json:"auth_time"`
	JTI      string `json:"jti"`
	ISS      string `json:"iss"`
	// AUD               string                 `json:"aud"`
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
	UserID            string                 `json:"-"`
	MapClaims         jwt.MapClaims          `json:"-"`
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

func (u *UserJWT) GetExpirationTime() (*jwt.NumericDate, error) {
	if u.EXP == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(u.EXP, 0)), nil
}

func (u *UserJWT) GetIssuedAt() (*jwt.NumericDate, error) {
	if u.IAT == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(u.IAT, 0)), nil
}

func (u *UserJWT) GetNotBefore() (*jwt.NumericDate, error) {
	if u.AuthTime == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(u.AuthTime, 0)), nil
}

func (u *UserJWT) GetAudience() (jwt.ClaimStrings, error) {
	if audVal, ok := u.MapClaims["aud"]; ok {
		switch aud := audVal.(type) {
		case string:
			return jwt.ClaimStrings{aud}, nil
		case []interface{}:
			var result jwt.ClaimStrings
			for _, v := range aud {
				if s, ok := v.(string); ok {
					result = append(result, s)
				}
			}
			return result, nil
		}
	}
	return nil, nil
}

func (u *UserJWT) GetIssuer() (string, error) {
	return u.ISS, nil
}

func (u *UserJWT) GetSubject() (string, error) {
	return u.UPN, nil
}

func (u *UserJWT) GetID() (string, error) {
	return u.JTI, nil
}

const UserAnonymous = "ANONYMOUS"

func IsRequestorAdmin(ctx context.Context, request *http.Request) (bool, error) {
	jwt, err := GetUserFromRequest(ctx, request)
	if err != nil {
		return false, err
	}
	return IsAdmin(jwt), nil
}

func IsAdmin(user *UserJWT) bool {
	if user.RealmAccess != nil {
		for _, role := range user.RealmAccess.Roles {
			if strings.EqualFold("DevSecOps", role) {
				return true
			}
		}
	}
	if user.Groups != nil {
		admin := DefaultEnvironment.Get("ADMIN_GROUP")

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

func GetUserTokenFromRequest(ctx context.Context, request *http.Request) (string, error) {
	tokens := request.Header["Authorization"]
	if len(tokens) == 1 {
		Info(ctx, "cloudy.GetUserTokenFromRequest Found Token in request header")
		return tokens[0], nil

	} else if len(tokens) > 1 {
		return "", Error(ctx, "cloudy.GetUserTokenFromRequest Multiple Tokens found in request header: %v\n", tokens)
	}

	token := request.URL.Query().Get("bearer")
	if token != "" {
		Info(ctx, "cloudy.GetUserTokenFromRequest Found Token in bearer")
		return token, nil
	}

	return "", Error(ctx, "cloudy.GetUserTokenFromRequest No Tokens found")
}

func GetUserFromRequest(ctx context.Context, request *http.Request) (*UserJWT, error) {

	token, err := GetUserTokenFromRequest(ctx, request)
	if err != nil {
		_ = Error(ctx, "cloudy.jwt.GetUserFromRequest error: %v", err)
		return nil, err
	}

	if token == "" {
		return nil, Error(ctx, "cloudy.jwt.GetUserFromRequest Empty token: %v\n", token)
	}

	return GetUserInfoFromToken(ctx, token), nil
}

// GetUserInfoFromToken Gets a user information from the JWT (Authorization Header)
func GetUserInfoFromToken(ctx context.Context, token string) *UserJWT {
	if token == "" || strings.EqualFold(token, "Bearer undefined") {
		_ = Error(ctx, "GetUserInfoFromToken Bearer token undefined")
		return &UserJWT{
			PreferredUserName: UserAnonymous,
			Email:             "None",
		}
	}

	claims, err := ParseToken(token)
	if err != nil {
		_ = Error(ctx, "GetUserInfoFromToken ParseToken Error %v", err)

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
		fmt.Printf("CUSTOM CLAIMS ERROR: %v\n", err)
		return nil, err
	}

	mapClaims := jwt.MapClaims(make(map[string]interface{}))
	_, _, err = parser.ParseUnverified(tokenToParse, &mapClaims)
	if err != nil {
		fmt.Printf("MAP CLAIMS ERROR: %v\n", err)
		return nil, err
	}

	// UPN and Email must be lower!
	claims.Email = strings.ToLower(claims.Email)
	claims.UPN = strings.ToLower(claims.UPN)

	if claims.UPN == "" {
		Info(context.Background(), "UPN not found in JWT (Email: %s)", claims.Email)
	}
	claims.MapClaims = mapClaims

	return &claims, nil
}

func EncodeJWT(user *UserJWT) string {
	hmacSampleSecret := []byte("This is a cool secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)

	// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
	// 		jwt.MapClaims{
	// 	   "foo": "bar",
	// 	   "nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	//    })

	tokenString, err := token.SignedString(hmacSampleSecret)
	if err != nil {
		panic(err)
	}

	return tokenString
}
