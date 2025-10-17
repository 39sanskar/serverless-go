package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"log" // For logging repository errors

	"github.com/39sanskar/serverless-go/pkg/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var (
	ErrorFailedToUnmarshalRecord  = "failed to unmarshal record"
	ErrorFailedToFetchRecord      = "failed to fetch record from DynamoDB"
	ErrorInvalidUserData          = "invalid user data"
	ErrorCouldNotMarshalItem      = "could not marshal item"
	ErrorCouldNotDeleteItem       = "could not delete item"
	ErrorCouldNotDynamoPutItem    = "could not put item into DynamoDB"
	ErrorUserAlreadyExists        = "user already exists"
	ErrorUserDoesNotExist         = "user does not exist"
	ErrorCouldNotScanItems        = "could not scan items from DynamoDB"
	ErrorInvalidLastEvaluatedKey  = "invalid last evaluated key for pagination"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	FetchUser(email string) (*models.User, error)
	FetchUsers(limit int, lastEvaluatedKey string) ([]models.User, string, error)
	CreateUser(user models.User) (*models.User, error)
	UpdateUser(user models.User) (*models.User, error)
	DeleteUser(email string) error
}

// DynamoDBUserRepository implements UserRepository for DynamoDB.
type DynamoDBUserRepository struct {
	client    dynamodbiface.DynamoDBAPI
	tableName string
}

// NewDynamoDBUserRepository creates a new DynamoDBUserRepository.
func NewDynamoDBUserRepository(client dynamodbiface.DynamoDBAPI, tableName string) *DynamoDBUserRepository {
	return &DynamoDBUserRepository{
		client:    client,
		tableName: tableName,
	}
}

// FetchUser retrieves a single user by email.
func (repo *DynamoDBUserRepository) FetchUser(email string) (*models.User, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(repo.tableName),
	}

	result, err := repo.client.GetItem(input)
	if err != nil {
		log.Printf("DynamoDB GetItem error: %v", err)
		return nil, fmt.Errorf("%s: %w", ErrorFailedToFetchRecord, err)
	}

	if result.Item == nil {
		return nil, nil // User not found
	}

	item := new(models.User)
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		log.Printf("DynamoDB UnmarshalMap error: %v", err)
		return nil, fmt.Errorf("%s: %w", ErrorFailedToUnmarshalRecord, err)
	}
	return item, nil
}

// FetchUsers retrieves multiple users with pagination.
// Returns a list of users, the last evaluated key for next page, and an error.
func (repo *DynamoDBUserRepository) FetchUsers(limit int, lastEvaluatedKey string) ([]models.User, string, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(repo.tableName),
		Limit:     aws.Int64(int64(limit)),
	}

	// Add ExclusiveStartKey for pagination if lastEvaluatedKey is provided
	if lastEvaluatedKey != "" {
		var startKey map[string]*dynamodb.AttributeValue
		err := json.Unmarshal([]byte(lastEvaluatedKey), &startKey)
		if err != nil {
			log.Printf("Invalid lastEvaluatedKey JSON: %v", err)
			return nil, "", errors.New(ErrorInvalidLastEvaluatedKey)
		}
		input.ExclusiveStartKey = startKey
	}

	result, err := repo.client.Scan(input)
	if err != nil {
		log.Printf("DynamoDB Scan error: %v", err)
		return nil, "", fmt.Errorf("%s: %w", ErrorCouldNotScanItems, err)
	}

	users := new([]models.User)
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, users)
	if err != nil {
		log.Printf("DynamoDB UnmarshalListOfMaps error: %v", err)
		return nil, "", fmt.Errorf("%s: %w", ErrorFailedToUnmarshalRecord, err)
	}

	// Marshal LastEvaluatedKey for the next page
	var newLastEvaluatedKey string
	if result.LastEvaluatedKey != nil {
		keyBytes, err := json.Marshal(result.LastEvaluatedKey)
		if err != nil {
			log.Printf("Error marshaling LastEvaluatedKey: %v", err)
			return nil, "", fmt.Errorf("could not marshal LastEvaluatedKey: %w", err)
		}
		newLastEvaluatedKey = string(keyBytes)
	}

	return *users, newLastEvaluatedKey, nil
}

// CreateUser creates a new user in DynamoDB.
func (repo *DynamoDBUserRepository) CreateUser(user models.User) (*models.User, error) {
	// Check if user already exists
	currentUser, err := repo.FetchUser(user.Email)
	if err != nil {
		return nil, err // Propagate original error
	}
	if currentUser != nil {
		return nil, errors.New(ErrorUserAlreadyExists)
	}

	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		log.Printf("DynamoDB MarshalMap error: %v", err)
		return nil, fmt.Errorf("%s: %w", ErrorCouldNotMarshalItem, err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(repo.tableName),
		// ConditionExpression: aws.String("attribute_not_exists(email)"), // Ensure user doesn't exist
	}

	_, err = repo.client.PutItem(input)
	if err != nil {
		log.Printf("DynamoDB PutItem error: %v", err)
		return nil, fmt.Errorf("%s: %w", ErrorCouldNotDynamoPutItem, err)
	}
	return &user, nil
}

// UpdateUser updates an existing user in DynamoDB.
func (repo *DynamoDBUserRepository) UpdateUser(user models.User) (*models.User, error) {
	// Check if user exists
	currentUser, err := repo.FetchUser(user.Email)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New(ErrorUserDoesNotExist)
	}

	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		log.Printf("DynamoDB MarshalMap error: %v", err)
		return nil, fmt.Errorf("%s: %w", ErrorCouldNotMarshalItem, err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(repo.tableName),
		// ConditionExpression: aws.String("attribute_exists(email)"), // Ensure user exists
	}

	_, err = repo.client.PutItem(input)
	if err != nil {
		log.Printf("DynamoDB PutItem error: %v", err)
		return nil, fmt.Errorf("%s: %w", ErrorCouldNotDynamoPutItem, err)
	}
	return &user, nil
}

// DeleteUser deletes a user by email from DynamoDB.
func (repo *DynamoDBUserRepository) DeleteUser(email string) error {
	// Check if user exists before attempting to delete
	currentUser, err := repo.FetchUser(email)
	if err != nil {
		return err
	}
	if currentUser == nil {
		return errors.New(ErrorUserDoesNotExist)
	}

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(repo.tableName),
	}
	_, err = repo.client.DeleteItem(input)
	if err != nil {
		log.Printf("DynamoDB DeleteItem error: %v", err)
		return fmt.Errorf("%s: %w", ErrorCouldNotDeleteItem, err)
	}
	return nil
}