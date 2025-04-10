package simplewire

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Users is a component which requires several other components to function.
// This example shows how to use simplewire when the dest is a struct.
// Components which have Users injected must use type *Users (as opposed to User)
type Users struct {
	Accounts    Accounts `inject:"accounts"`
	DB          Database `inject:"db"`
	initialized bool
}

// Accounts is an interface, perhaps giving opportunity for Mock behavior in tests.
// We will use it as example when the dest is an interface.
// Components which have Accounts injected must use type Accounts (as opposed to *Accounts)
type Accounts interface {
	AccountsByID(accountID string) (*Account, error)
	AccountsByUser(username string) ([]*Account, error)
}

// AccountsS is a struct that implements Accounts.  It also has some dependencies which need to be injected.
type AccountsS struct {
	Users *Users   `inject:"users"`
	DB    Database `inject:"db"`
}

// Database is an interface that will be injected into both services.  A mock implementation is provided for the tests.
type Database interface {
	UserByID(userID string) (*User, error)
	UserByUsername(username string) (*User, error)
	AccountByID(accountID string) (*Account, error)
	AccountsByUserID(userID string) ([]*Account, error)
}

// Components is the struct used as a reference of things to inject
type Components struct {
	Users    *Users
	Accounts Accounts
	DB       Database
}

// TestConnect tests that the various links between each component are properly set up.
func TestConnect(t *testing.T) {
	components := Components{
		Users:    &Users{},
		Accounts: &AccountsS{},
		DB:       &MockDB{},
	}
	_, err := Connect(components)
	assert.NoError(t, err)

	assert.True(t, components.Users.initialized, "components.Users should have had the Init function called")
	assert.Same(t, components.DB, components.Users.DB, "components.Users should have been wired with a pointer to components.DB")
	assert.Same(t, components.Accounts, components.Users.Accounts, "components.Users should have been wired with a pointer to components.Accounts")
	accountsS := components.Accounts.(*AccountsS)
	assert.Same(t, components.Users, accountsS.Users, "components.Accounts should have been wired with a pointer to components.Users")
	assert.Same(t, components.DB, accountsS.DB, "components.Accounts should have been wired with a pointer to components.DB")
}

// TestInject tests the injector returned by Connect can be used to inject more things later.
func TestInject(t *testing.T) {
	components := Components{
		Users:    &Users{},
		Accounts: &AccountsS{},
		DB:       &MockDB{},
	}

	// Connect our components
	injector, err := Connect(components)
	assert.NoError(t, err)

	type Thing struct {
		// Users is a struct so we need to use a pointer
		Users *Users `inject:"users"`
		// Accounts is an interface, so we should not use a pointer
		Accounts Accounts `inject:"accounts"`
	}

	t1 := Thing{}
	err = injector.Inject(&t1)
	assert.NoError(t, err)
	assert.Same(t, components.Users, t1.Users, "t1 should have been wired with a pointer to components.Users")
	assert.Same(t, components.Accounts, t1.Accounts, "t1 should have been wired with a pointer to components.Accounts")

	// find the accounts for a given user ->
	accounts, err := t1.Accounts.AccountsByUser(testUsername)
	assert.NoError(t, err)
	assert.Len(t, accounts, 1)
	assert.Equal(t, testAccountID, accounts[0].AccountID)
}

type User struct {
	UserID   string
	Username string
}

type Account struct {
	AccountID string
	UserID    string
}

type MockDB struct{}

var (
	_ Initializable = &Users{}
	_ Accounts      = AccountsS{}
	_ Database      = MockDB{}
)

const (
	testUsername  = "the-username"
	testUserID    = "the-user-id"
	testAccountID = "the-account-id"
)

func (u *Users) Init() error {
	u.initialized = true
	return nil
}

func (u Users) GetUser(username string) (*User, error) {
	return u.DB.UserByUsername(username)
}

func (u Users) FromAccountID(accountID string) (*User, error) {
	sess, err := u.Accounts.AccountsByID(accountID)
	if err != nil {
		return nil, err
	}
	return u.DB.UserByID(sess.UserID)
}

func (s AccountsS) AccountsByUser(username string) ([]*Account, error) {
	user, err := s.Users.GetUser(username)
	if err != nil {
		return nil, err
	}
	return s.DB.AccountsByUserID(user.UserID)
}

func (s AccountsS) AccountsByID(accountID string) (*Account, error) {
	return s.DB.AccountByID(accountID)
}

func (d MockDB) UserByID(userID string) (*User, error) {
	user := mockUser()
	if userID == user.UserID {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (d MockDB) UserByUsername(username string) (*User, error) {
	user := mockUser()
	if username == user.Username {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (d MockDB) AccountsByUserID(userID string) ([]*Account, error) {
	account := mockAccount()
	if account.UserID == userID {
		return []*Account{account}, nil
	}
	return nil, errors.New("user not found")
}

func (d MockDB) AccountByID(accountID string) (*Account, error) {
	account := mockAccount()
	if account.AccountID == accountID {
		return account, nil
	}
	return nil, errors.New("account not found")
}

func mockUser() *User {
	return &User{
		UserID:   testUserID,
		Username: testUsername,
	}
}

func mockAccount() *Account {
	return &Account{
		AccountID: testAccountID,
		UserID:    testUserID,
	}
}
