package validators

import (
	"errors"
	"regexp"

	"github.com/39sanskar/serverless-go/pkg/models"
)

// Regex for email validation (a commonly used robust pattern)
var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// IsEmailValid checks if the provided email string is a valid email address.
func IsEmailValid(email string) bool {
	if len(email) < 3 || len(email) > 254 || !rxEmail.MatchString(email) {
		return false
	}
	return true
}

// ValidateUser performs comprehensive validation for a User struct.
func ValidateUser(user models.User) error {
	if user.Email == "" {
		return errors.New("email is required")
	}
	if !IsEmailValid(user.Email) {
		return errors.New("invalid email format")
	}
	if user.FirstName == "" {
		return errors.New("first name is required")
	}
	if user.LastName == "" {
		return errors.New("last name is required")
	}
	// Add more validation rules as needed (e.g., length, alphanumeric, etc.)
	return nil
}