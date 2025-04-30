// file:buzz/mod/m_bus/bus_auth/user.go

package bus_auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/rskv-p/buzz/mod/m_bus/bus_middle"
	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
)

//---------------------
// MIDDLEWARE: AUTHENTICATION
//---------------------

// AuthenticateUser is a middleware that authenticates users via username/password and checks permissions.
func AuthenticateUser() typ.IMiddleware {
	return bus_middle.NewMiddleware(func(req typ.IRequest) error {
		// Extract the headers from the request and handle any error
		headers, err := req.Headers()
		if err != nil {
			x_log.Error("Failed to get headers from request", err)
			return fmt.Errorf("failed to get headers: %v", err)
		}

		// Extract the username from the headers
		username, err := headers.Get("Username")
		if err != nil || username == "" {
			x_log.Error("Missing or invalid Username in headers")
			return fmt.Errorf("username is missing or invalid")
		}

		// Extract the password from the headers
		password, err := headers.Get("Password")
		if err != nil || password == "" {
			x_log.Error("Missing or invalid Password in headers")
			return fmt.Errorf("password is missing or invalid")
		}

		// Check user credentials in the database
		var user User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			x_log.Error("User not found in database", username)
			return fmt.Errorf("invalid username or password")
		}

		// Compare the hashed password with the provided one
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			x_log.Error("Invalid credentials, incorrect password", username)
			return fmt.Errorf("invalid username or password")
		}

		// Attach the authenticated user to the request for further processing (optional)
		req.SetHeader("authenticated_user", username)
		x_log.Info("User authenticated successfully", username)

		return nil
	})
}

//---------------------
// MIDDLEWARE: TOPIC PERMISSIONS
//---------------------

// CheckTopicPermissions is a middleware that checks if the user has permission to access the topic.
func CheckTopicPermissions() typ.IMiddleware {
	return bus_middle.NewMiddleware(func(req typ.IRequest) error {
		// Retrieve the headers from the request and handle any error
		headers, err := req.Headers()
		if err != nil {
			x_log.Error("Failed to get headers from request", err)
			return fmt.Errorf("failed to get headers: %v", err)
		}

		// Retrieve the authenticated user from the headers
		username, _ := headers.Get("authenticated_user")
		if username == "" {
			x_log.Error("No authenticated user found in headers")
			return fmt.Errorf("no authenticated user found")
		}

		// Fetch the user details from the database (preload permissions)
		var user User
		if err := db.Preload("Permissions").Where("username = ?", username).First(&user).Error; err != nil {
			x_log.Error("User not found in database", username)
			return fmt.Errorf("user %s not found", username)
		}

		// Extract the topic from the request
		topic := req.GetSubject()
		if topic == "" {
			x_log.Error("Topic is missing from request", username)
			return fmt.Errorf("topic is missing")
		}

		// Check if the user has permission for the topic
		for _, permission := range user.Permissions {
			if WildcardMatch(topic, permission.Name) {
				x_log.Info("User has permission to access topic", username, topic)
				return nil
			}
		}

		// If no permission matches, return an error
		x_log.Error("User does not have permission to access topic", username, topic)
		return fmt.Errorf("user %s does not have permission to access topic %s", username, topic)
	})
}
