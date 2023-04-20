package helpers

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/cristalhq/jwt/v4"
)

func CheckToken(token string) (string, error) {
	var key string
	if os.Getenv("JWT_SECRET") != "" {
		key = os.Getenv("JWT_SECRET")
	} else {
		key = "secret"
	}

	verifier, err := jwt.NewVerifierHS(jwt.HS512, []byte(key))
	if err != nil {
		return "", err
	}

	tokenBytes := []byte(token)
	newToken, err := jwt.Parse(tokenBytes, verifier)
	if err != nil {
		return "", err
	}

	err = verifier.Verify(newToken)
	if err != nil {
		return "", err
	}

	// get Registered claims
	var newClaims jwt.RegisteredClaims
	err = json.Unmarshal(newToken.Claims(), &newClaims)
	if err != nil {
		return "", err
	}

	err = jwt.ParseClaims(tokenBytes, verifier, &newClaims)
	if err != nil {
		return "", err
	}

	if !newClaims.IsValidAt(time.Now()) {
		return "", errors.New("invalid time")
	}

	return newClaims.Subject, nil
}
