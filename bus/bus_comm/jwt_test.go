// file: buz/bus/bus_comm/jwt_test.go

package bus_comm_test

import (
	"testing"
	"time"

	"github.com/rskv-p/buzz/bus/bus_comm"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJWT(t *testing.T) {
	// Test data
	subject := "test-user"
	secretKey := "test-secret-key"
	expiration := 1 * time.Hour

	// Generate JWT token
	token, err := bus_comm.GenerateJWT(subject, secretKey, expiration)

	// Assert no error occurred and token is not empty
	assert.NoError(t, err, "Expected no error while generating JWT")
	assert.NotEmpty(t, token, "Expected non-empty token")
}

func TestVerifyJWT(t *testing.T) {
	// Test data
	subject := "test-user"
	secretKey := "test-secret-key"
	expiration := 1 * time.Hour

	// Generate JWT token
	token, err := bus_comm.GenerateJWT(subject, secretKey, expiration)
	assert.NoError(t, err, "Expected no error while generating JWT")

	// Verify the generated token
	claims, err := bus_comm.VerifyJWT(token, secretKey)
	assert.NoError(t, err, "Expected no error while verifying JWT")

	// Assert the claims are correct
	assert.Equal(t, subject, (*claims)["sub"], "Expected subject to match")
}

func TestVerifyInvalidJWT(t *testing.T) {
	// Test data
	invalidToken := "invalid-token"
	secretKey := "test-secret-key"

	// Attempt to verify an invalid token
	_, err := bus_comm.VerifyJWT(invalidToken, secretKey)
	assert.Error(t, err, "Expected error while verifying invalid JWT")
}
