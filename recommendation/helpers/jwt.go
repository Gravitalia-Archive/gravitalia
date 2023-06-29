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

// CheckToken allows to check the authenticity of a token
// and return the user vanity if it is a real token
func Check(token string) (string, error) {
	block, _ := pem.Decode([]byte(os.Getenv("RSA_PUBLIC_KEY")))
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
