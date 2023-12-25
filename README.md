# Exo

Attempt to port [ecto](https://hexdocs.pm/ecto) Elixir's library to the Go ecosystem!

## Example

```go
package main

import "regexp"
import "errors"
import "github.com/zoedsoupe/exo"

type User struct {
    Username string
    Password string
}

func main() {
    var user User

    // could be a body from a HTTP request
    attrs := map[string]interface{}{
        "Username": "foo",
        "Password": "123",
        "Don't Exist": "super value"
    }

    // here I want all fields, but you can set it mannually
    fields := exo.Fields(user)
    changeset := exo.New(user, attrs) // creates a changeset
        .Cast(fields) // filter fields and validate types
        .ValidateLength("Username", 3) // validate the specific lenth
        .ValidateLength("Password", 10)
        .ValidateChange("Password", validatePassword)

    user, err := exo.Apply[User](changeset)
    if err != nil {
      // here will be a list of changeset errors OR
      // an error on the struct building step
      // example: mismatching types
      panic(error)
    }

    username := user.Username // "foo"
    password := user.Password // "password"
}

func validatePassword(field string, curr interface{}) (bool, error) {
    switch c := curr.(type) {
      case string:
        re := regexp.Mustcompile("[a-z]|[A-Z]")
        if re.FindString(c.(string)) == "" {
          return false, errors.New("password must match alphabetic chars")
        }

        return true, nil
        
  		default:
  			return false, errors.New("Field isn't a string")
     }
}
```
