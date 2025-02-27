# URL Shortener Lambda Function URL with Go

This project implements a serverless URL shortener using AWS Lambda Function URL with Go runtime (Amazon Linux 2023) and DynamoDB.

## Features

- Create shortened URLs with a random 5-character code
- Redirect to original URLs using the short code
- Serverless architecture using AWS Lambda Function URL (no API Gateway needed)
- DynamoDB for persistent storage
- Go implementation with AWS Lambda Go runtime (AL2023)

## Prerequisites

- AWS CLI configured with appropriate permissions
- Go 1.21 or later
- Make

## Project Structure

- `main.go` - Go code for the Lambda function
- `template.yaml` - CloudFormation template to deploy all necessary resources
- `Makefile` - Helper commands for building and deploying
- `go.mod` and `go.sum` - Go module dependencies

## Setup Instructions

1. Clone this repository

2. Update the `S3_BUCKET` variable in the Makefile to a unique bucket name for your deployment

3. Build and deploy the application:

```bash
make deploy
```

This will:
- Build the Go binary for Lambda
- Create an S3 bucket if it doesn't exist
- Upload the Lambda package to S3
- Deploy the CloudFormation stack with all resources

4. After deployment, the URL of your API will be displayed. It will look something like:
```
https://abcdef123456.lambda-url.us-east-1.on.aws/
```

## API Usage

### Create a Short URL

```bash
curl -X POST https://your-lambda-url.on.aws/url -H "Content-Type: application/json" -d '{"url":"https://example.com/very/long/url/that/needs/shortening"}'
```

Response:
```json
{
  "originalUrl": "https://example.com/very/long/url/that/needs/shortening",
  "shortUrl": "https://your-lambda-url.on.aws/url/ab1Cd"
}
```

### Use a Short URL

Simply visit the short URL in a browser or make a GET request to it:

```bash
curl -L https://your-lambda-url.on.aws/url/ab1Cd
```

This will redirect to the original URL.

## Customization

- Change the short code length by modifying the `codeLength` constant in `main.go`
- Adjust the lambda timeout, memory size, or other properties in `template.yaml`
- Modify CORS settings in the Lambda Function URL resource

## Go Modules

The application requires the following Go modules:

```bash
go get github.com/aws/aws-lambda-go/events
go get github.com/aws/aws-lambda-go/lambda
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue
go get github.com/aws/aws-sdk-go-v2/service/dynamodb
```

## Cleanup

To remove all resources created by this project:

```bash
aws cloudformation delete-stack --stack-name url-shortener
```

## Security Considerations

- The Lambda Function URL is publicly accessible by default
- Consider adding authentication (change `AuthType` to `AWS_IAM` in the CloudFormation template)
- Add rate limiting to prevent abuse
- Implement URL validation to prevent malicious URLs