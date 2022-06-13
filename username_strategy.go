package cloudy

import (
	"context"
	"fmt"
	"strings"

	"github.com/appliedres/cloudy/models"
)

type DuplicateUserMode int

const UserAny DuplicateUserMode = 1
const UserExists DuplicateUserMode = 2
const UserDoesNotExist DuplicateUserMode = 3

type UsernameStrategy interface {
	// GuessUsername provides the best username for a given user object. The pattern should contain
	GuessUsername(ctx context.Context, user *models.User, mode DuplicateUserMode, manager UserManager) (string, error)
}

type FirstDotLastUsernameStrategy struct{}

func (strat *FirstDotLastUsernameStrategy) GuessUsername(ctx context.Context, user *models.User, mode DuplicateUserMode, manager UserManager, userDomain string) (string, error) {
	username := user.FirstName + "." + user.LastName
	return FindMatchingUser(ctx, username, mode, manager, userDomain)
}

type FirstInitialLastUsernameStrategy struct{}

func (strat *FirstInitialLastUsernameStrategy) GuessUsername(ctx context.Context, user *models.User, mode DuplicateUserMode, manager UserManager, userDomain string) (string, error) {
	username := user.FirstName[:1] + user.LastName
	return FindMatchingUser(ctx, username, mode, manager, userDomain)
}

func FindMatchingUser(ctx context.Context, username string, mode DuplicateUserMode, manager UserManager, userDomain string) (string, error) {

	guess := addDomain(username, userDomain)
	if mode == UserAny {
		return guess, nil
	}

	for count := 1; count < 100; count++ {
		if count > 1 {
			guess = addDomain(fmt.Sprintf("%v.%v", username, count), userDomain)
		}

		user, err := manager.GetUser(ctx, guess)
		if err != nil {
			return username, nil
		}

		if mode == UserExists && user != nil {
			return username, nil
		}
		if mode == UserExists && user == nil {
			return "", nil
		}
		if mode == UserDoesNotExist && user == nil {
			return username, nil
		}

		Info(ctx, "Found User %v", user.DisplayName)
	}
	return "", nil
}

func addDomain(user string, domain string) string {
	if domain != "" {
		return user + "@" + domain
	}
	return user
}

func FixName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "")
	return name
}
