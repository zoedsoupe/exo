# Exo

Attempt to port [ecto](https://hexdocs.pm/ecto) Elixir's library to the Go ecosystem!

## Example

```go
package main

import (
	"errors"
	"fmt"
	"regexp"
	"zoedsoupe/exo"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string
	Password string
}

func main() {
	var user User

	// could be a body from a HTTP request
	attrs := map[string]interface{}{
		"Username":    "foo",
		"Password":    "1234567890",
		"Don't Exist": "super value",
	}

	// here I want all fields, but you can set it mannually
	fields := []string{"Username", "Password"}
	re := regexp.MustCompile("[0-9]")

	changeset := exo.Cast[User](attrs) // creates new changeset
	    .ValidateChange("Username", LengthValidator{Min: 3, Max: 3}) // validate the exact string length
	    .ValidateChange("Username", DeniedValidator{Username: "denied"}) // validate custom changes
	    .ValidateChange("Password", LengthValidator{Min: 10})
	    .ValidateChange("Password", FormatValidator{Pattern: re}) // validate using regex
	    .UpdateChange("Password", hashPassword) // transform changes into the changeset

	user, err := exo.Apply[User](changeset)
	if err != nil {
		// here will be a list of changeset errors OR
		// an error on the struct building step
		// example: mismatching types
		panic(err)
	}

	username := user.Username // "foo"
	password := user.Password // "password"
	fmt.Println(username, password)
}

func hashPassword(pass interface{}) interface{} {
	s := pass.(string)
	b := []byte(s)
	hash, _ := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost)
	return string(hash)
}

type DeniedValidator struct {
	Username string
}

func (dv DeniedValidator) Validate(_ string, username interface{}) (bool, error) {
	if username == dv.Username {
		return false, fmt.Errorf("the username %s is denied", dv.Username)
	}

	return true, nil
}
```
