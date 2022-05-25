package account

import (
	"context"
	"fmt"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/datastore"
	"github.com/appliedres/cloudy/models"
)

// An AccountManager is a composite manager that handles the use case of having an identity
// store (like AWS cognito) and a locally managed user database for more enhanced information
// This manager helps tie both together.
type AccountManager[T any] struct {

	// The Datatype that is managing the additional information for the account
	DT datastore.Datatype[T]

	// The User Manager (third-party) that is responsible for the management of the user
	Users cloudy.UserManager

	// Function that retrieves the account it from the user object (like the email or upn)
	GetAccountId func(ctx context.Context, User *models.User) string

	// Function that can take an account and generate / augment all the necessary information
	// in the user. For instance. If the primary place to store the user name is in the Account
	// then this method would copy the name from the account to the user when it changes.
	// that way the user account is always correct.
	// The bool return indicated if any changes were made that need to be saved
	SyncUserToAccount func(ctx context.Context, Account *T, User *models.User) (bool, error)

	// Opposite of the SyncUserToAccount function
	// The bool return indicated if any changes were made that need to be saved
	SyncAccountToUser func(ctx context.Context, Account *T, User *models.User) (bool, error)

	// Determines if this manager should create missing accounts. If the missing account creation
	// is enabled then when a user is found without an account then the account will be created,
	// synced with the user and then saved
	CreateMissingAccount bool

	// Opposite of the SyncUserToAccount function
	// The bool return indicated if any changes were made that need to be saved
	NewAccount func(ctx context.Context, User *models.User) (*T, error)

	// OnEnable (Optional) Called after the user manager enables a user, this is a hook
	// to do extra processing on the account
	OnEnable func(ctx context.Context, Account *T, User *models.User) error

	// OnDisable (Optional) Called after the user manager disables a user, this is a hook
	// to do extra processing on the account
	OnDisable func(ctx context.Context, Account *T, User *models.User) error
}

type UserAccount[T any] struct {
	Account *T
	User    *models.User
}

func (acm *AccountManager[T]) GetAccount(ctx context.Context, user *models.User) (account *T, err error) {
	// Get the field from the user object
	aid := acm.GetAccountId(ctx, user)

	// No Account
	if aid == "" {
		err = fmt.Errorf("no account id can be derived from user")
		return
	}

	// Try to lookup the account
	account, err = acm.DT.Get(ctx, aid)

	// No account, but we should make one
	if account == nil && acm.CreateMissingAccount {
		// Create this account
		account, err = acm.NewAccount(ctx, user)
		if err != nil {
			return
		}

		// save this account
		account, err = acm.DT.Save(ctx, account)
	}
	return
}

// Retrieves a specific user account.
func (acm *AccountManager[T]) Get(ctx context.Context, uid string) (*UserAccount[T], error) {
	// Get the user from the system
	user, err := acm.Users.GetUser(ctx, uid)
	if err != nil {
		return nil, err
	}

	ua := &UserAccount[T]{
		User: user,
	}

	// Try to lookup the account
	acct, err := acm.GetAccount(ctx, user)
	if err != nil {
		return ua, err
	}

	// No account, but we should make one
	if acct == nil && acm.CreateMissingAccount {
		// Create this account
		acct, err = acm.NewAccount(ctx, user)
		if err != nil {
			return ua, err
		}

		// save this account
		acct, err = acm.DT.Save(ctx, acct)
		if err != nil {
			return ua, err
		}
	}

	ua.Account = acct
	return ua, nil
}

// Creates a new User account. This is created in the Data store and the third party system
func (acm *AccountManager[T]) New(ctx context.Context, newUser *UserAccount[T]) (*UserAccount[T], error) {
	newu, err := acm.Users.NewUser(ctx, newUser.User)
	if err != nil {
		return nil, err
	}

	ua := &UserAccount[T]{
		User: newu,
	}

	newa := newUser.Account
	if newa == nil {

		// Create this account
		newa, err = acm.NewAccount(ctx, newu)
		if err != nil {
			return ua, err
		}
	}

	newa, err = acm.DT.Save(ctx, newUser.Account)
	if err != nil {
		return ua, err
	}
	ua.Account = newa
	return ua, nil
}

func (acm *AccountManager[T]) Enable(ctx context.Context, uid string) error {
	user, err := acm.Users.GetUser(ctx, uid)
	if err != nil {
		return err
	}

	err = acm.Users.Disable(ctx, uid)
	if err != nil {
		return err
	}

	if acm.OnEnable != nil {
		acct, err := acm.GetAccount(ctx, user)
		if err != nil {
			return err
		}

		err = acm.OnEnable(ctx, acct, user)
		if err != nil {
			return err
		}
	}
	return nil
}

func (acm *AccountManager[T]) Disable(ctx context.Context, uid string) error {
	user, err := acm.Users.GetUser(ctx, uid)
	if err != nil {
		return err
	}

	err = acm.Users.Disable(ctx, uid)
	if err != nil {
		return err
	}

	if acm.OnEnable != nil {
		acct, err := acm.GetAccount(ctx, user)
		if err != nil {
			return err
		}

		err = acm.OnDisable(ctx, acct, user)
		if err != nil {
			return err
		}
	}
	return nil
}

func (acm *AccountManager[T]) Delete(ctx context.Context, uid string) (bool, error) {
	user, err := acm.Users.GetUser(ctx, uid)
	if err != nil {
		return false, err
	}
	err = acm.Users.DeleteUser(ctx, uid)
	if err != nil {
		return false, err
	}

	aid := acm.GetAccountId(ctx, user)
	if aid != "" {
		err = acm.DT.Delete(ctx, aid)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}
