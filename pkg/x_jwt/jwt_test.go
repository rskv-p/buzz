// file: buzz/pkg/x_jwt/jwt_test.go

package x_jwt_test

import (
	"testing"
	"time"

	"github.com/rskv-p/buzz/pkg/x_init"
	"github.com/rskv-p/buzz/pkg/x_jwt"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJWT(t *testing.T) {
	x_init.Init()

	// Test data
	subject := "test-user"
	secretKey := "test-secret-key"
	expiration := 1 * time.Hour

	// Generate JWT token
	token, err := x_jwt.GenerateJWT(subject, secretKey, expiration)

	// Assert no error occurred and token is not empty
	assert.NoError(t, err, "Expected no error while generating JWT")
	assert.NotEmpty(t, token, "Expected non-empty token")
}

func TestVerifyJWT(t *testing.T) {
	x_init.Init()

	// Test data
	subject := "test-user"
	secretKey := "test-secret-key"
	expiration := 1 * time.Hour

	// Generate JWT token
	token, err := x_jwt.GenerateJWT(subject, secretKey, expiration)
	assert.NoError(t, err, "Expected no error while generating JWT")

	// Verify the generated token
	claims, err := x_jwt.VerifyJWT(token, secretKey)
	assert.NoError(t, err, "Expected no error while verifying JWT")

	// Assert the claims are correct
	assert.Equal(t, subject, (*claims)["sub"], "Expected subject to match")
}

func TestVerifyInvalidJWT(t *testing.T) {
	x_init.Init()

	// Test data
	invalidToken := "invalid-token"
	secretKey := "test-secret-key"

	// Attempt to verify an invalid token
	_, err := x_jwt.VerifyJWT(invalidToken, secretKey)
	assert.Error(t, err, "Expected error while verifying invalid JWT")
}
