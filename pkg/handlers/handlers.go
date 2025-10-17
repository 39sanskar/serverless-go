package handlers

import (
	"encoding/json"
	"net/http"
	"strconv" // For pagination

	"github.com/39sanskar/serverless-go/pkg/models" // Use models package for User struct
	"github.com/39sanskar/serverless-go/pkg/repository"
	"github.com/39sanskar/serverless-go/pkg/validators"
	"github.com/aws/aws-lambda-go/events"
)

// UserHandler provides methods for handling user-related API requests.
type UserHandler struct {
	userRepo repository.UserRepository
}

// NewUserHandler creates a new UserHandler instance.
func NewUserHandler(userRepo repository.UserRepository) UserHandler {
	return UserHandler{
		userRepo: userRepo,
	}
}

// GetUser handles GET requests for users.
// It can fetch a single user by email or all users with pagination.
func (h *UserHandler) GetUser(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	email := req.QueryStringParameters["email"]

	if email != "" {
		// Fetch single user
		user, err := h.userRepo.FetchUser(email)
		if err != nil {
			return apiResponse(http.StatusBadRequest, ErrorBody{
				ErrorMsg: StringPtr(err.Error()),
			})
		}
		if user == nil {
			return apiResponse(http.StatusNotFound, ErrorBody{
				ErrorMsg: StringPtr("User not found"),
			})
		}
		return apiResponse(http.StatusOK, user)
	}

	// Fetch all users with optional pagination
	limitStr := req.QueryStringParameters["limit"]
	lastEvaluatedKey := req.QueryStringParameters["lastEvaluatedKey"] // For pagination token

	limit := 10 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	users, newLastEvaluatedKey, err := h.userRepo.FetchUsers(limit, lastEvaluatedKey)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			ErrorMsg: StringPtr(err.Error()),
		})
	}

	responseBody := map[string]interface{}{
		"users": users,
	}
	if newLastEvaluatedKey != "" {
		responseBody["lastEvaluatedKey"] = newLastEvaluatedKey
	}

	return apiResponse(http.StatusOK, responseBody)
}

// CreateUser handles POST requests to create a new user.
func (h *UserHandler) CreateUser(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var user models.User
	if err := json.Unmarshal([]byte(req.Body), &user); err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			ErrorMsg: StringPtr("Invalid request body"),
		})
	}

	// Validate user data
	if err := validators.ValidateUser(user); err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			ErrorMsg: StringPtr(err.Error()),
		})
	}

	createdUser, err := h.userRepo.CreateUser(user)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			ErrorMsg: StringPtr(err.Error()),
		})
	}
	return apiResponse(http.StatusCreated, createdUser)
}

// UpdateUser handles PUT requests to update an existing user.
func (h *UserHandler) UpdateUser(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var user models.User
	if err := json.Unmarshal([]byte(req.Body), &user); err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			ErrorMsg: StringPtr("Invalid request body"),
		})
	}

	// Email is required for update
	if user.Email == "" {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			ErrorMsg: StringPtr("Email is required for user update"),
		})
	}

	// Validate user data (excluding email format if not changing, but general content validation)
	// For simplicity, re-validating the whole user struct.
	if err := validators.ValidateUser(user); err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			ErrorMsg: StringPtr(err.Error()),
		})
	}

	updatedUser, err := h.userRepo.UpdateUser(user)
	if err != nil {
		// Specific error checks for 404 vs 400
		if err.Error() == repository.ErrorUserDoesNotExist {
			return apiResponse(http.StatusNotFound, ErrorBody{
				ErrorMsg: StringPtr("User not found for update"),
			})
		}
		return apiResponse(http.StatusBadRequest, ErrorBody{
			ErrorMsg: StringPtr(err.Error()),
		})
	}
	return apiResponse(http.StatusOK, updatedUser)
}

// DeleteUser handles DELETE requests to delete a user by email.
func (h *UserHandler) DeleteUser(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	email := req.QueryStringParameters["email"]
	if email == "" {
		return apiResponse(http.StatusBadRequest, ErrorBody{
			ErrorMsg: StringPtr("Email query parameter is required for deletion"),
		})
	}

	err := h.userRepo.DeleteUser(email)
	if err != nil {
		// Specific error checks for 404 vs 400
		if err.Error() == repository.ErrorUserDoesNotExist {
			return apiResponse(http.StatusNotFound, ErrorBody{
				ErrorMsg: StringPtr("User not found for deletion"),
			})
		}
		return apiResponse(http.StatusBadRequest, ErrorBody{
			ErrorMsg: StringPtr(err.Error()),
		})
	}
	return apiResponse(http.StatusNoContent, nil) // 204 No Content for successful deletion
}