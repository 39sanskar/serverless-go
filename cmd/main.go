package main

import (
	"log"

	"github.com/39sanskar/serverless-go/config"
	"github.com/39sanskar/serverless-go/pkg/handlers"
	"github.com/39sanskar/serverless-go/pkg/repository"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// Declare dynaClient globally for direct use, or pass it via a handler struct if preferred for strict DI.
// For AWS Lambda, initializing it once outside the handler function is a common and efficient pattern.
var dynamoClient *dynamodb.DynamoDB
var userHandler handlers.UserHandler

func init() {
	// Initialize configurations from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize AWS session
	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWSRegion),
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err) // Use log.Fatalf instead of panic
	}

	// Initialize DynamoDB client
	dynamoClient = dynamodb.New(awsSession)

	// Initialize the user repository and handler
	userRepo := repository.NewDynamoDBUserRepository(dynamoClient, cfg.TableName)
	userHandler = handlers.NewUserHandler(userRepo)
}

func main() {
	lambda.Start(handler)
}

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Add logging for incoming requests
	log.Printf("Received request: %s %s", req.HTTPMethod, req.Path)

	switch req.HTTPMethod {
	case "GET":
		return userHandler.GetUser(req)
	case "POST":
		return userHandler.CreateUser(req)
	case "PUT":
		return userHandler.UpdateUser(req)
	case "DELETE":
		return userHandler.DeleteUser(req)
	default:
		return handlers.UnhandledMethod()
	}
}

