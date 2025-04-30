// file:buzz/mod/m_bus/bus_auth/service.go

package bus_auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rskv-p/buzz/mod/m_bus/bus_middle"
	"github.com/rskv-p/buzz/pkg/x_jwt"
	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
)

//---------------------
// UTILITY FUNCTIONS
//---------------------

// logErrorWithContext logs an error along with the action and service name for better debugging.
func logErrorWithContext(action, serviceName string, err error) {
	x_log.Error("Error occurred while", action, "for service", serviceName, ":", err)
}

//---------------------
// JWT GENERATION & AUTHENTICATION
//---------------------

// GenerateServiceJWT generates a JWT token for a service account with the given service name and secret key.
func GenerateServiceJWT(serviceName, secretKey string, expiration time.Duration) (string, error) {
	x_log.Info("Generating JWT for service", serviceName, "with expiration time of", expiration)

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
		x_log.Error("Error generating JWT for service", serviceName, ":", err)
		return "", fmt.Errorf("error generating JWT: %v", err)
	}

	x_log.Info("JWT for service", serviceName, "generated successfully")
	return tokenString, nil
}

//---------------------
// MIDDLEWARE: AUTHENTICATION & PERMISSIONS
//---------------------

// AuthenticateService is a middleware that authenticates a service via a token (e.g., JWT) and checks permissions.
func AuthenticateService(secretKey string) typ.IMiddleware {
	return bus_middle.NewMiddleware(func(req typ.IRequest) error {
		// Extract headers from the request and handle any error
		headers, err := req.Headers()
		if err != nil {
			x_log.Error("Failed to retrieve headers from request:", err)
			return fmt.Errorf("failed to get headers: %v", err)
		}

		// Extract token from the headers
		token, err := headers.Get("Token")
		if err != nil || token == "" {
			x_log.Error("Missing or invalid Token in headers")
			return fmt.Errorf("token is missing or invalid")
		}

		// Log the token processing (do not log the full token in production)
		x_log.Debug("Processing token for service authentication, token length:", len(token))

		// Verify the token
		claims, err := x_jwt.VerifyJWT(token, secretKey)
		if err != nil {
			x_log.Error("Invalid token:", err)
			return fmt.Errorf("invalid token: %v", err)
		}

		// Extract 'sub' claim (subject)
		mapClaims := *claims
		subject, ok := mapClaims["sub"].(string)
		if !ok {
			x_log.Error("Token does not contain 'sub' claim or it's not a string")
			return fmt.Errorf("invalid token: missing or invalid 'sub' claim")
		}

		// Attach the authenticated service to the request
		req.SetHeader("authenticated_service", subject)
		x_log.Info("Service", subject, "authenticated successfully with token")

		return nil
	})
}

// CheckServiceTopicPermissions is a middleware that checks if the service has permission to access the topic.
func CheckServiceTopicPermissions() typ.IMiddleware {
	return bus_middle.NewMiddleware(func(req typ.IRequest) error {
		// Extract headers from the request and handle any error
		headers, err := req.Headers()
		if err != nil {
			x_log.Error("Failed to retrieve headers from request:", err)
			return fmt.Errorf("failed to get headers: %v", err)
		}

		// Retrieve the authenticated service from the headers
		serviceName, _ := headers.Get("authenticated_service")
		if serviceName == "" {
			x_log.Error("No authenticated service found in headers")
			return fmt.Errorf("no authenticated service found")
		}

		// Log the service being processed
		x_log.Debug("Checking permissions for service:", serviceName)

		// Fetch service account from the database
		var serviceAccount ServiceAccount
		if err := db.Where("service_name = ?", serviceName).First(&serviceAccount).Error; err != nil {
			logErrorWithContext("fetching service from DB", serviceName, err)
			return fmt.Errorf("service %s not found in database", serviceName)
		}

		// Extract topic from the request
		topic := req.GetSubject()
		if topic == "" {
			x_log.Error("Topic is missing from the request for service", serviceName)
			return fmt.Errorf("topic is missing")
		}

		// Check if the service has permission for the topic
		for _, permission := range serviceAccount.Permissions {
			if WildcardMatch(topic, permission.Name) {
				x_log.Info("Service", serviceName, "has permission to access topic", topic)
				return nil
			}
		}

		// If no permission matches, log and return an error
		x_log.Error("Service", serviceName, "does not have permission to access topic", topic)
		return fmt.Errorf("service %s does not have permission to access topic %s", serviceName, topic)
	})
}

//---------------------
// WILDCARD MATCHING
//---------------------

// WildcardMatch checks if a topic matches a permission with wildcards.
func WildcardMatch(topic, permission string) bool {
	x_log.Debug("Matching topic", topic, "with permission", permission)
	topicParts := strings.Split(topic, ".")
	permissionParts := strings.Split(permission, ".")

	for i, part := range permissionParts {
		if part == "*" {
			x_log.Debug("Wildcard '*' matches part of the topic", topicParts[i])
			continue
		} else if part == ">" {
			x_log.Debug("Wildcard '>' matches the rest of the topic", topic)
			return true
		} else if i >= len(topicParts) || part != topicParts[i] {
			x_log.Debug("No match for part", part, "in topic", topic)
			return false
		}
	}
	x_log.Debug("Match found for topic", topic, "and permission", permission)
	return len(topicParts) >= len(permissionParts)
}
