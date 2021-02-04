# simplewire 

Simplewire is a lightweight dependency injection framework for Go (golang) that uses struct tags and requires no additional build steps.

## Why

Dependency injection is a thing.  It can be complicated to explain, so I'll leave it to [Wikipedia](https://en.wikipedia.org/wiki/Dependency_injection).  Other go modules I've found work differently and require extra build steps and produce generated code.  This one uses reflection and is only used to set fields within your structs.  I feel there is less magic in the code that is produced and that using struct tags is more idiomatic.

## Basic Usage

Use a struct tag to name a set of dependencies you will inject.  Something like `service`, `component`, or `provider` could make sense.  You can choose the key of the struct tag to fit your use case.

As an example,

```go
type Users struct {
  // Accounts needs to be injected into Users
  Accounts    *Accounts `service:"accounts"`
}

func (u Users) GetBalance(userID string) int {
  ...
  // Once injected, we can just use its API
  return u.Accounts.GetBalance(user.accountID)
}

type Accounts struct {
  // Users needs to be injected into Accounts
  Users    *Users `service:"users"`
}

func (a Accounts) GetOwner(accountID string) User {
  ...
  return a.Users.GetUser(account.userID)
}

```

We also need to create a list of services to use as the reference.  Essentially this is just a dictionary of things we want to inject.  When we find a struct tag we are looking for, the value of the tag will be matched to the names of the field in the reference.  The matching is not case sensitive.

```go

type Services struct {
  Users *User
  Accounts *Accounts
}

func main() {
  // Define the set of things that can be injected.
  services := Services{
    Users:    &Users{},
    Accounts: &Accounts{},
  }

  injector, err := simplewire.Connect("service", services)
  if err != nil {
    panic(err)
  }

  // After running connect, Users and Accounts are now wired up!  If you are
  // done, you may not need to keep the Injector around. 

  // If you want to use the set of dependencies on more clients which cannot be 
  // used as providers, use the Injector returned by simplewire.Connect.
  handler := WebHandler{}
  err = injector.Inject(&handler)
  if err != nil {
    panic(err)
  }
}
```

## Further reading

For now, all I have to offer is the [test file](./simplewire_test.go), which might be helpful as an example if you want comment or use this module. 