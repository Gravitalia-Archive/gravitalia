package helpers

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/cristalhq/jwt/v5"
)

// Check allow to check the token authenticity
func Check(token string) (string, error) {
	var key string
	if os.Getenv("JWT_SECRET") != "" {
		key = os.Getenv("JWT_SECRET")
	} else {
		key = "secret"
	}

	hsverifier, err := jwt.NewVerifierHS(jwt.HS512, []byte(key))
	if err != nil {
		return "", err
	}

	parsedToken, err := jwt.Parse([]byte(token), hsverifier)
	if err != nil {
		return "", err
	}

	err = hsverifier.Verify(parsedToken)
	if err != nil {
		return "", err
	}

	var claims jwt.RegisteredClaims
	err = json.Unmarshal(parsedToken.Claims(), &claims)
	if err != nil {
		return "", err
	}

	err = jwt.ParseClaims([]byte(token), hsverifier, &claims)
	if err != nil {
		return "", err
	}

	if !claims.IsValidAt(time.Now()) {
		return "", errors.New("invalid time")
	}

	return claims.Subject, nil
}
