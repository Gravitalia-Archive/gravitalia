package main

import (
	"regexp"
	"testing"

	"github.com/Gravitalia/gravitalia/helpers"
)

func TestCreateToken(t *testing.T) {
	jwt, err := helpers.CreateToken("test")
	if err != nil {
		t.Fatalf(`CreateToken("test") = %q, got an error`, err)
	}

	jwtCheck := regexp.MustCompile(`(^[A-Za-z0-9-_]*\.[A-Za-z0-9-_]*\.[A-Za-z0-9-_]*$)`)
	if !jwtCheck.MatchString(jwt) {
		t.Fatalf(`CreateToken("test") = %q, want match for %#q, nil`, jwt, jwtCheck)
	}
}
