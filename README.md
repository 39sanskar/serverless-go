# Serverless Go User API

This project demonstrates a serverless API built with Go, AWS Lambda, and Amazon DynamoDB. It provides basic CRUD (Create, Read, Update, Delete) operations for user management. The code is structured for maintainability, testability, and scalability following industry best practices.

## Features

*   **CRUD Operations:** Create, retrieve (single/all with pagination), update, and delete users.
*   **AWS Lambda Integration:** Designed to run seamlessly as an AWS Lambda function.
*   **Amazon DynamoDB Persistence:** Stores user data in a NoSQL DynamoDB table.
*   **Clean Architecture Principles:**
    *   **Configuration Management:** Centralized loading of environment variables.
    *   **Separation of Concerns:** Clear distinction between handlers, models, repositories, and validators.
    *   **Dependency Injection:** Handlers depend on interfaces (repositories) for easier testing and flexibility.
*   **Robust Error Handling:** Granular error messages and appropriate HTTP status codes.
*   **Input Validation:** Server-side validation for user data.
*   **Pagination:** Supports fetching a limited number of users with a `lastEvaluatedKey` for subsequent pages.
*   **Structured Logging:** Basic logging for request tracing and error diagnostics.

## Project Structure
```bash
serverless-go/
├── build/ # Compiled deployment artifacts
│ └── main.zip # Zipped Go Lambda function (for deployment)
├── cmd/ # Main application entry point
│ └── main.go # Lambda handler and initialization
├── config/ # Application configuration loading
│ └── config.go # Loads environment variables
├── pkg/ # Core application logic
│ ├── handlers/ # API Gateway event handlers
│ │ ├── api_response.go # Generic API response utility
│ │ └── handlers.go # User API methods (GetUser, CreateUser, etc.)
│ ├── models/ # Data structures (e.g., User struct)
│ │ └── user.go # User model definition
│ ├── repository/ # Data access layer interface and implementations
│ │ └── user_repository.go # DynamoDB user repository implementation
│ └── validators/ # Input validation logic
│ └── validators.go # Functions for validating user input
├── go.mod # Go module definition
├── go.sum # Go module checksums
└── README.md # This documentation file
```


## Setup and Deployment

* Go (version 1.18 or higher)
* AWS CLI configured with appropriate permissions
* An AWS account
* Serverless Framework (recommended for easy deployment) or AWS SAM CLI

## 1. Initialize Go Module

```bash
go mod init github.com/39sanskar/serverless-go # Replace with your Go module path
go mod tidy
```

## 2. Create DynamoDB Table

* You need a DynamoDB table named LambdaInGoUser (or whatever you configure DYNAMODB_TABLE_NAME to be) with email as the primary key.

* Using AWS CLI

```bash
aws dynamodb create-table \
    --table-name LambdaInGoUser \
    --attribute-definitions \
        AttributeName=email,AttributeType=S \
    --key-schema \
        AttributeName=email,KeyType=HASH \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --region <your-aws-region>
```

## 3. Deployment using Serverless Framework (Recommended)

* Create a serverless.yml file in the root of your project

```yaml
service: serverless-go-user-api

provider:
  name: aws
  runtime: go1.x
  region: us-east-1 # Change to your preferred AWS region
  memorySize: 128
  timeout: 10
  environment:
    AWS_REGION: ${self:provider.region}
    DYNAMODB_TABLE_NAME: LambdaInGoUser # Ensure this matches your table name
  iamRoleStatements:
    - Effect: Allow
      Action:
        - dynamodb:GetItem
        - dynamodb:PutItem
        - dynamodb:UpdateItem
        - dynamodb:DeleteItem
        - dynamodb:Scan
      Resource: "arn:aws:dynamodb:${self:provider.region}:*:table/${self:environment.DYNAMODB_TABLE_NAME}"

package:
  patterns:
    - '!./**'
    - './bin/**' # Include compiled binary

functions:
  userApi:
    handler: bin/main # Path to your compiled Go binary
    events:
      - http:
          path: users
          method: any
          cors: true # Enable CORS for API Gateway
```

## Build and Deploy
```bash
# Compile your Go application for AWS Lambda
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/main cmd/main.go

# Deploy using Serverless Framework
serverless deploy

# Build using Command 
zip -j build/main.zip build/main
```
* After deployment, the serverless deploy command will output the API Gateway endpoint URL.

## 4. Deployment using AWS SAM (Alternative)

* f you prefer AWS SAM, you'll need a template.yaml and a build process.

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31
Description: A serverless Go API for user management.

Globals:
  Function:
    Runtime: go1.x
    MemorySize: 128
    Timeout: 10
    Environment:
      Variables:
        AWS_REGION: !Ref AWS::Region
        DYNAMODB_TABLE_NAME: LambdaInGoUser # Ensure this matches your table name
    Policies:
      - DynamoDBCrudPolicy:
          TableName: !Ref UserTable

Resources:
  UserApiFunction:
    Type: AWS::Serverless::Function
    Properties:
      Handler: main
      CodeUri: s3://YOUR_S3_BUCKET/PATH_TO_YOUR_ZIP_FILE # Or use sam build locally
      Events:
        CatchAll:
          Type: Api
          Properties:
            Path: /users/{proxy+}
            Method: ANY
            Cors:
              AllowOrigin: "'*'"
              AllowHeaders: "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token'"
              AllowMethods: "'GET,PUT,POST,DELETE,OPTIONS'"

  UserTable:
    Type: AWS::DynamoDB::Table
    Properties:
      TableName: LambdaInGoUser
      AttributeDefinitions:
        - AttributeName: email
          AttributeType: S
      KeySchema:
        - AttributeName: email
          KeyType: HASH
      ProvisionedThroughput:
        ReadCapacityUnits: 5
        WriteCapacityUnits: 5

Outputs:
  UserApi:
    Description: "API Gateway endpoint URL for Prod stage for User service"
    Value: !Sub "https://${ServerlessRestApi}.execute-api.${AWS::Region}.amazonaws.com/Prod/users/"
```

## Build and Deploy (SAM)
```bash
# Build the Go application locally
sam build

# Deploy the SAM application
sam deploy --guided
```

## API Endpoints

* All endpoints are relative to your API Gateway URL (e.g., https://xxxxxx.execute-api.us-east-1.amazonaws.com/Prod/users).

### 1. Create User (POST)
• Endpoint: /users
• Method: POST
• Request Body (JSON):

```json
{
    "email": "test@example.com",
    "firstName": "John",
    "lastName": "Doe"
}
```
• Response (201 Created):
```json
{
    "email": "test@example.com",
    "firstName": "John",
    "lastName": "Doe"
}
```
• Error Responses:
• 400 Bad Request: If request body is invalid, data validation fails (e.g., invalid email, missing fields), or user with that email already exists.

### 2. Get User(s) (GET)
• Endpoint: /users
• Method: GET

Get Single User by Email
• Query Parameters: email=<user-email> (e.g., /users?email=test@example.com)

• Response (200 OK):
```json
{
    "email": "test@example.com",
    "firstName": "John",
    "lastName": "Doe"
}
```

• Error responses:
• 400 Bad Request: If there's an issue fetching from the database.
• 404 Not Found: If the user with the specified email does not exist.

• Get All Users (with Pagination)

• Query Parameters (Optional)
• limit=<number>: Maximum number of users to return (default: 10).
• lastEvaluatedKey=<json-string>: The LastEvaluatedKey from a previous response to fetch the next page.

• Response (200 OK)
```json
{
    "users": [
        {
            "email": "user1@example.com",
            "firstName": "Alice",
            "lastName": "Smith"
        },
        {
            "email": "user2@example.com",
            "firstName": "Bob",
            "lastName": "Johnson"
        }
    ],
    "lastEvaluatedKey": "{\"email\":{\"S\":\"user2@example.com\"}}" # Present if more items are available
}
```
• Error Responses:
• 400 Bad Request: If lastEvaluatedKey is malformed or other database issues.

### 3. Update User (PUT)
• Endpoint: /users
• Method: PUT
• Request Body (JSON):
```json
{
    "email": "test@example.com",
    "firstName": "Jonathan",
    "lastName": "Davis"
}
```
• Note: email is required in the body to identify the user.

• Response (200 OK)
```json
{
    "email": "test@example.com",
    "firstName": "Jonathan",
    "lastName": "Davis"
}
```
• Error Responses:
• 400 Bad Request: If request body is invalid, data validation fails, or email is missing.
• 404 Not Found: If the user with the specified email does not exist.

### 4. Delete User(DELETE)
• Endpoint: /users

• Method: DELETE

• Query Parameters: email=<user-email> (e.g., /users?email=test@example.com)

• Response (204 No Content): (No body on successful deletion)

• Error Responses:

• 400 Bad Request: If email query parameter is missing or other database issues.

• 404 Not Found: If the user with the specified email does not exist.

