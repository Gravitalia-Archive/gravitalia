package helpers

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cristalhq/jwt/v5"
)

// CreateToken allows to create JWT tokens
func CreateToken(vanity string) (string, error) {
	block, _ := pem.Decode([]byte(os.Getenv("RSA_PRIVATE_KEY")))
	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

	fmt.Println(key)

	signer, err := jwt.NewSignerRS(jwt.RS256, key)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()

	token, err := jwt.NewBuilder(signer).Build(&jwt.RegisteredClaims{
		Subject:   vanity,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Date(now.Year(), now.Month(), now.Day()+7, now.Hour(), now.Minute(), 0, 0, time.UTC)),
		Issuer:    "https://www.gravitalia.com",
	})
	if err != nil {
		return "", err
	}

	return token.String(), nil
}

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
