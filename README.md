# Exo

[![CI](https://github.com/zoedsoupe/exo/actions/workflows/ci.yml/badge.svg)](https://github.com/zoedsoupe/exo/actions/workflows/ci.yml)

Attempt to port [ecto](https://hexdocs.pm/ecto) Elixir's library to the Go ecosystem!

## Example

```go
package main

import (
	"errors"
	"fmt"
	"regexp"
	"github.com/zoedsoupe/exo/changeset"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string
	Password string
}

func main() {
	// could be a body from a HTTP request
	attrs := map[string]interface{}{
		"Username":    "foo",
		"Password":    "1234567890",
		"Don't Exist": "super value",
	}

	// here I want all fields, but you can set it mannually
	fields := []string{"Username", "Password"}
	re := regexp.MustCompile("[0-9]")

	c := changeset.Cast[User](attrs) // creates new changeset
	    .ValidateChange("Username", LengthValidator{Min: 3, Max: 3}) // validate the exact string length
	    .ValidateChange("Username", DeniedValidator{Username: "denied"}) // validate custom changes
	    .ValidateChange("Password", LengthValidator{Min: 10})
	    .ValidateChange("Password", FormatValidator{Pattern: re}) // validate using regex
	    .UpdateChange("Password", hashPassword) // transform changes into the changeset

	// here we're creating a new instance of `User`
	// if you want to update an existing user
	// use: changeset.Apply[User](&existingUser, changeset)
	err := changeset.ApplyNew[User](changeset)
	if err != nil {
		// here will be the invalid changeset itself
		// it implements the error interface so you can
		// call `Error` method or even the convinience
		// `ErrorJSON` for HTTP server responses
		panic(err)
	}

	username := user.Username // "foo"
	password := user.Password // "password"
	fmt.Println(username, password)
}

func hashPassword(pass interface{}) (interface{}, error) {
	s := pass.(string)
	b := []byte(s)
	hash, err := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost)
	if err != {
		return nil, err // it will added to the changeset
	}

	return string(hash), nil
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
