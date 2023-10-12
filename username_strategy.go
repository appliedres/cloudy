package cloudy

import (
	"context"
	"fmt"
	"strings"

	"github.com/appliedres/cloudy/models"
)

type DuplicateUserMode int

const UserAny DuplicateUserMode = 0
const UserExists DuplicateUserMode = 1
const UserDoesNotExist DuplicateUserMode = 2

var DuplicateUserNames = [3]string{"UserAny", "UserExists", "UserDoesNotExist"}

type UsernameStrategy interface {
	// GuessUsername provides the best username for a given user object. The pattern should contain
	GuessUsername(ctx context.Context, user *models.User, mode DuplicateUserMode, manager UserManager) (string, error)
}

type FirstDotLastUsernameStrategy struct{}

func (strat *FirstDotLastUsernameStrategy) GuessUsername(ctx context.Context, user *models.User, mode DuplicateUserMode, manager UserManager, userDomain string) (string, error) {
	username := strings.ToLower(user.FirstName + "." + user.LastName)
	return FindMatchingUser(ctx, username, mode, manager, userDomain)
}

type FirstInitialLastUsernameStrategy struct{}

func (strat *FirstInitialLastUsernameStrategy) GuessUsername(ctx context.Context, user *models.User, mode DuplicateUserMode, manager UserManager, userDomain string) (string, error) {
	username := strings.ToLower(user.FirstName[:1] + user.LastName)
	return FindMatchingUser(ctx, username, mode, manager, userDomain)
}

func FindMatchingUser(ctx context.Context, username string, mode DuplicateUserMode, manager UserManager, userDomain string) (string, error) {

	lowerUser := strings.ToLower(username)
	lowerDomain := strings.ToLower(userDomain)
	guess := AddDomain(lowerUser, lowerDomain)

	if mode == UserAny {
		return guess, nil
	}

	for count := 1; count < 20; count++ {
		if count > 1 {
			guess = AddDomain(fmt.Sprintf("%v.%v", lowerUser, count), lowerDomain)
		}

		user, err := manager.GetUser(ctx, guess)
		if err != nil {
			return "", Error(ctx, "Error getting user during FindMatchingUser: %v", err)
		}

		if mode == UserExists && user != nil {
			Info(ctx, "UserExists -> guessing username %s", guess)
			return guess, nil
		}
		if mode == UserDoesNotExist && user == nil {
			Info(ctx, "UserDoesNotExist -> guessing username %s", guess)
			return guess, nil
		}
	}

	return "", Error(ctx, "Unable to find a matching user for username: %s (mode: %s)", lowerUser, DuplicateUserNames[mode])
}

func AddDomain(user string, domain string) string {
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
