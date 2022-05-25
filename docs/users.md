# Users

There are 2 primary modes of operation. The first is where the user store is entirely handled by a third party (like Cognito). The second mode is where the basic identity is handled by a third party but the primary information is handled by a JSON Data store. This second model is used when extesive account managagement is used. For instance a `User` object paired with an `Account` object. The `Account` object is the one that the system mostly interacts with. 
In the second model the caller has to determine the join behavior. For instance, is the email the unique id for a user in the system or is it a user id.

```go
umgr := NewCognitoUserManager(...)
amgr := AccountManager[MyAccount](umgr, {})
```


NewAccountManager[MyAccount](NewAzureUserManager(), {})

Generalized to the interceptor pattern

type BeforeAction


type InteceptedService[] interface {
  // Before Errors CANCEL the action
  Before(ctx context, *T item) (*T, error) 

  // After Errors are up to the caller to decide what to do
  After(ctx context, *T item) (*T, error)
}



## Models

may consider the interface pattern... 
User
- GetX()
- SetX()
(one per field? or use standard attributes? IDK)
----------
Serialization utils
- GetSource()
- ToJson()
----------
In Source: 
- FromJson()

and then concrete structs
PROS: 
- Can add data in concrete struct
- Can cast to concrete struct
CONS:
- Clumsy

Alternates:
- Map interface with named fields
- Less type consistency
- could add validators
- need toJWTClaims or something
- PITA for some types

single, concrete object
User:
    type: object
    properties: 
      ID:
        type: string
      UserName:
        type: string
      FirstName:
        type: string
      LastName:
        type: string
      Email:
        type: string
      Company: 
        type: string
      MobilePhone:
        type: string
      OfficePhone:
        type: string
      Department:
        type: string
      JobTitle:
        type: string
      Extra: <-- MAP[any]any
        type: object

## Interface


## Examples


## Azure Configuration


## AWS Configuration