package cloudy

var CredentialSources = make(map[string]CredentialSource)

// Manages groups that users are part of.This can be seperate
// from the user manager or it can be the same.
type CredentialSource interface {
	ReadFromEnv(env *Environment) interface{}
}
