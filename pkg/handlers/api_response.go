package handlers

import (
	"encoding/json"
	"log" // Added for logging errors during JSON marshaling
  "net/http"
	"github.com/aws/aws-lambda-go/events"
)

// ErrorBody represents a standardized error response structure.
type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty"`
}

// apiResponse creates a standardized APIGatewayProxyResponse.
func apiResponse(status int, body interface{}) (*events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers: map[string]string{"Content-Type": "application/json"},
	}
	resp.StatusCode = status

	// Marshal the body to JSON. Handle potential errors during marshaling.
	stringBody, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error marshaling response body: %v", err)
		// Fallback to a generic error message if the original body couldn't be marshaled
		errorJson, _ := json.Marshal(ErrorBody{ErrorMsg: StringPtr("Failed to marshal response body")})
		resp.Body = string(errorJson)
		resp.StatusCode = 500 // Internal Server Error
		return &resp, nil
	}

	resp.Body = string(stringBody)
	return &resp, nil
}

// UnhandledMethod returns a 405 Method Not Allowed response.
func UnhandledMethod() (*events.APIGatewayProxyResponse, error) {
	return apiResponse(http.StatusMethodNotAllowed, ErrorBody{ErrorMsg: StringPtr("Method Not Allowed")})
}

// Helper to get a pointer to a string.
func StringPtr(s string) *string {
	return &s
}