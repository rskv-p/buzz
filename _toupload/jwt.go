// file: buzz/pkg/x_jwt/jwt.go

package x_jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

//-------------------------------------------------
// GenerateJWT - Generates a JWT token with the given subject and secret key
//-------------------------------------------------

// GenerateJWT generates a JWT token with the given subject and secret key, and sets the expiration time.
func GenerateJWT(subject, secretKey string, expiration time.Duration) (string, error) {
	//Infof("Generating JWT for subject: %s with expiration: %v", subject, expiration)

	// Create the claims for the JWT
	claims := &jwt.RegisteredClaims{
		Subject:   subject,                                        // Subject (usually user ID or username)
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)), // Set expiration time
		Issuer:    "bus_system",                                   // Optional: Issuer of the token
	}

	// Create a new token with HMAC signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token using the secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		//Errorf("Error generating JWT for subject %s: %v", subject, err)
		return "", err
	}

	//Infof("Successfully generated JWT for subject: %s", subject)
	return tokenString, nil
}

//-------------------------------------------------
// VerifyJWT - Verifies the JWT token and returns the claims if valid
//-------------------------------------------------

// VerifyJWT verifies the JWT token and returns the claims if the token is valid.
func VerifyJWT(tokenStr string, secret string) (*jwt.MapClaims, error) {
	//Infof("Verifying JWT token")

	// Parse and validate the token using the provided secret key
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			//Errorf("Unexpected signing method: %v", t.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		//		Errorf("Error parsing JWT token: %v", err)
		return nil, err
	}

	// Check if the token claims are valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		//		Infof("Successfully verified JWT token")
		return &claims, nil
	}

	//	Errorf("Invalid JWT token")
	return nil, fmt.Errorf("invalid token")
}
