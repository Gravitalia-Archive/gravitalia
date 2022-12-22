package helpers

import (
	"os"
	"time"

	"github.com/cristalhq/jwt/v4"
)

// CreateToken allows to create JWT tokens
func CreateToken(vanity string) (string, error) {
	signer, err := jwt.NewSignerHS(jwt.HS512, []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()

	token, err := jwt.NewBuilder(signer).Build(&jwt.RegisteredClaims{
		Subject:   vanity,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Date(now.Year(), now.Month()+1, now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)),
		Issuer:    "https://www.gravitalia.com",
	})
	if err != nil {
		return "", err
	}

	return token.String(), nil
}
