// file: buz/bus/bus_auth/user.go

package bus_auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/rskv-p/buzz/bus/bus_comm"
	"github.com/rskv-p/buzz/bus/bus_middle"
	"github.com/rskv-p/buzz/bus/bus_type"
)

//---------------------
// MIDDLEWARE: AUTHENTICATION
//---------------------

// AuthenticateUser is a middleware that authenticates users via username/password and checks permissions.
func AuthenticateUser() bus_type.IMiddleware {
	return bus_middle.NewMiddleware(func(req bus_type.IRequest) error {
		// Extract the headers from the request and handle any error
		headers, err := req.Headers()
		if err != nil {
			bus_comm.Errorf("Failed to get headers from request: %v", err)
			return fmt.Errorf("failed to get headers: %v", err)
		}

		// Extract the username from the headers
		username, err := headers.Get("Username")
		if err != nil || username == "" {
			bus_comm.Errorf("Missing or invalid Username in headers")
			return fmt.Errorf("username is missing or invalid")
		}

		// Extract the password from the headers
		password, err := headers.Get("Password")
		if err != nil || password == "" {
			bus_comm.Errorf("Missing or invalid Password in headers")
			return fmt.Errorf("password is missing or invalid")
		}

		// Check user credentials in the database
		var user User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			bus_comm.Errorf("User %s not found in database", username)
			return fmt.Errorf("invalid username or password")
		}

		// Compare the hashed password with the provided one
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			bus_comm.Errorf("Invalid credentials for user %s: incorrect password", username)
			return fmt.Errorf("invalid username or password")
		}

		// Attach the authenticated user to the request for further processing (optional)
		req.SetHeader("authenticated_user", username)
		bus_comm.Infof("User %s authenticated successfully", username)

		return nil
	})
}

//---------------------
// MIDDLEWARE: TOPIC PERMISSIONS
//---------------------

// CheckTopicPermissions is a middleware that checks if the user has permission to access the topic.
func CheckTopicPermissions() bus_type.IMiddleware {
	return bus_middle.NewMiddleware(func(req bus_type.IRequest) error {
		// Retrieve the headers from the request and handle any error
		headers, err := req.Headers()
		if err != nil {
			bus_comm.Errorf("Failed to get headers from request: %v", err)
			return fmt.Errorf("failed to get headers: %v", err)
		}

		// Retrieve the authenticated user from the headers
		username, _ := headers.Get("authenticated_user")
		if username == "" {
			bus_comm.Errorf("No authenticated user found in headers")
			return fmt.Errorf("no authenticated user found")
		}

		// Fetch the user details from the database (preload permissions)
		var user User
		if err := db.Preload("Permissions").Where("username = ?", username).First(&user).Error; err != nil {
			bus_comm.Errorf("User %s not found in database", username)
			return fmt.Errorf("user %s not found", username)
		}

		// Extract the topic from the request
		topic := req.GetSubject()
		if topic == "" {
			bus_comm.Errorf("Topic is missing from request for user %s", username)
			return fmt.Errorf("topic is missing")
		}

		// Check if the user has permission for the topic
		for _, permission := range user.Permissions {
			if WildcardMatch(topic, permission.Name) {
				bus_comm.Infof("User %s has permission to access topic %s", username, topic)
				return nil
			}
		}

		// If no permission matches, return an error
		bus_comm.Errorf("User %s does not have permission to access topic %s", username, topic)
		return fmt.Errorf("user %s does not have permission to access topic %s", username, topic)
	})
}
