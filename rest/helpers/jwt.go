package helpers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"time"

	"github.com/cristalhq/jwt/v5"
)

// CreateToken allows to create JWT tokens
func CreateToken(vanity string) (string, error) {
	var rsa_privaye_key string
	if os.Getenv("RSA_PRIVATE_KEY") == "" {
		rsa_privaye_key = ""
	} else {
		rsa_privaye_key = os.Getenv("RSA_PRIVATE_KEY")
	}
	block, _ := pem.Decode([]byte(rsa_privaye_key))
	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

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
	var rsa_public_key string
	if os.Getenv("RSA_PUBLIC_KEY") == "" {
		rsa_public_key = ""
	} else {
		rsa_public_key = os.Getenv("RSA_PUBLIC_KEY")
	}

	block, _ := pem.Decode([]byte(rsa_public_key))
	key, _ := x509.ParsePKIXPublicKey(block.Bytes)

	verifier, err := jwt.NewVerifierRS(jwt.RS256, key.(*rsa.PublicKey))
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
