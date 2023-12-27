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

	changeset := exo.New(user, attrs) // creates new changeset
	    .Cast(fields) // parse attrs fields based on type
	    .ValidateLength("Username", 3) // validate the exact string length
	    .ValidateChange("Username", blockDeniedUsername) // validate custom changes
	    .ValidateLength("Password", 10)
	    .ValidateFormat("Password", re) // validate using regex
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

func blockDeniedUsername(_ string, username interface{}) (bool, error) {
	if username == "denied" {
		return false, errors.New("the 'denied' username is, well, DENIED")
	}

	return true, nil
}
```
