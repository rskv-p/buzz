// file: buz/bus/bus_auth/service.go

package bus_auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rskv-p/buzz/bus/bus_comm"
	"github.com/rskv-p/buzz/bus/bus_middle"
	"github.com/rskv-p/buzz/bus/bus_type"
)

//---------------------
// UTILITY FUNCTIONS
//---------------------

// logErrorWithContext logs an error along with the action and service name for better debugging.
func logErrorWithContext(action, serviceName string, err error) {
	bus_comm.Errorf("Error occurred while %s for service %s: %v", action, serviceName, err)
}

//---------------------
// JWT GENERATION & AUTHENTICATION
//---------------------

// GenerateServiceJWT generates a JWT token for a service account with the given service name and secret key.
func GenerateServiceJWT(serviceName, secretKey string, expiration time.Duration) (string, error) {
	bus_comm.Infof("Generating JWT for service: %s with expiration time of %s", serviceName, expiration)

	// Create claims
	claims := &jwt.RegisteredClaims{
		Subject:   serviceName,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		Issuer:    "bus_system",
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		bus_comm.Errorf("Error generating JWT for service %s: %v", serviceName, err)
		return "", fmt.Errorf("error generating JWT: %v", err)
	}

	bus_comm.Infof("JWT for service %s generated successfully", serviceName)
	return tokenString, nil
}

//---------------------
// MIDDLEWARE: AUTHENTICATION & PERMISSIONS
//---------------------

// AuthenticateService is a middleware that authenticates a service via a token (e.g., JWT) and checks permissions.
func AuthenticateService(secretKey string) bus_type.IMiddleware {
	return bus_middle.NewMiddleware(func(req bus_type.IRequest) error {
		// Extract headers from the request and handle any error
		headers, err := req.Headers()
		if err != nil {
			bus_comm.Errorf("Failed to retrieve headers from request: %v", err)
			return fmt.Errorf("failed to get headers: %v", err)
		}

		// Extract token from the headers
		token, err := headers.Get("Token")
		if err != nil || token == "" {
			bus_comm.Errorf("Missing or invalid Token in headers")
			return fmt.Errorf("token is missing or invalid")
		}

		// Log the token processing (do not log the full token in production)
		if bus_comm.IsDebugEnabled() {
			bus_comm.Debugf("Processing token for service authentication, token length: %d", len(token))
		}

		// Verify the token
		claims, err := bus_comm.VerifyJWT(token, secretKey)
		if err != nil {
			bus_comm.Errorf("Invalid token: %v", err)
			return fmt.Errorf("invalid token: %v", err)
		}

		// Extract 'sub' claim (subject)
		mapClaims := *claims
		subject, ok := mapClaims["sub"].(string)
		if !ok {
			bus_comm.Errorf("Token does not contain 'sub' claim or it's not a string")
			return fmt.Errorf("invalid token: missing or invalid 'sub' claim")
		}

		// Attach the authenticated service to the request
		req.SetHeader("authenticated_service", subject)
		bus_comm.Infof("Service %s authenticated successfully with token", subject)

		return nil
	})
}

// CheckServiceTopicPermissions is a middleware that checks if the service has permission to access the topic.
func CheckServiceTopicPermissions() bus_type.IMiddleware {
	return bus_middle.NewMiddleware(func(req bus_type.IRequest) error {
		// Extract headers from the request and handle any error
		headers, err := req.Headers()
		if err != nil {
			bus_comm.Errorf("Failed to retrieve headers from request: %v", err)
			return fmt.Errorf("failed to get headers: %v", err)
		}

		// Retrieve the authenticated service from the headers
		serviceName, _ := headers.Get("authenticated_service")
		if serviceName == "" {
			bus_comm.Errorf("No authenticated service found in headers")
			return fmt.Errorf("no authenticated service found")
		}

		// Log the service being processed
		bus_comm.Debugf("Checking permissions for service: %s", serviceName)

		// Fetch service account from the database
		var serviceAccount ServiceAccount
		if err := db.Where("service_name = ?", serviceName).First(&serviceAccount).Error; err != nil {
			logErrorWithContext("fetching service from DB", serviceName, err)
			return fmt.Errorf("service %s not found in database", serviceName)
		}

		// Extract topic from the request
		topic := req.GetSubject()
		if topic == "" {
			bus_comm.Errorf("Topic is missing from the request for service %s", serviceName)
			return fmt.Errorf("topic is missing")
		}

		// Check if the service has permission for the topic
		for _, permission := range serviceAccount.Permissions {
			if WildcardMatch(topic, permission.Name) {
				bus_comm.Infof("Service %s has permission to access topic %s", serviceName, topic)
				return nil
			}
		}

		// If no permission matches, log and return an error
		bus_comm.Errorf("Service %s does not have permission to access topic %s", serviceName, topic)
		return fmt.Errorf("service %s does not have permission to access topic %s", serviceName, topic)
	})
}

//---------------------
// WILDCARD MATCHING
//---------------------

// WildcardMatch checks if a topic matches a permission with wildcards.
func WildcardMatch(topic, permission string) bool {
	bus_comm.Debugf("Matching topic '%s' with permission '%s'", topic, permission)
	topicParts := strings.Split(topic, ".")
	permissionParts := strings.Split(permission, ".")

	for i, part := range permissionParts {
		if part == "*" {
			bus_comm.Debugf("Wildcard '*' matches part of the topic '%s'", topicParts[i])
			continue
		} else if part == ">" {
			bus_comm.Debugf("Wildcard '>' matches the rest of the topic '%s'", topic)
			return true
		} else if i >= len(topicParts) || part != topicParts[i] {
			bus_comm.Debugf("No match for part '%s' in topic '%s'", part, topic)
			return false
		}
	}
	bus_comm.Debugf("Match found for topic '%s' and permission '%s'", topic, permission)
	return len(topicParts) >= len(permissionParts)
}
