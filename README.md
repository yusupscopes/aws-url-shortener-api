# URL Shortener Lambda Function URL with Go

This project implements a serverless URL shortener using AWS Lambda Function URL with Go runtime (Amazon Linux 2023) and DynamoDB.

## Features
![Image](https://github.com/user-attachments/assets/617bb323-fa57-498d-a109-bd6b9df07f32)
- Create shortened URLs with a random 5-character code
- Redirect to original URLs using the short code
- View analytics for shortened URLs
- Optional URL expiration with DynamoDB TTL
- Serverless architecture using AWS Lambda Function URL (no API Gateway needed)
- DynamoDB for persistent storage
- Go implementation with AWS Lambda Go runtime (AL2023)

## Architecture Overview

This URL shortener uses a serverless architecture with the following components:

- **AWS Lambda with Function URL**: Handles all HTTP requests without requiring API Gateway
- **DynamoDB**: Stores URL mappings with TTL support for expiration
- **CloudWatch Logs**: Captures logs for monitoring and debugging
- **CloudFormation**: Manages all infrastructure as code

## Prerequisites

- AWS CLI configured with appropriate permissions
- Go 1.21 or later
- Make

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
curl -X POST https://your-lambda-url.on.aws/shorten -H "Content-Type: application/json" -d '{"url":"https://example.com/very/long/url/that/needs/shortening", "expire_in_days": 7}'
```

Response:
```json
{  
  "short_url": "https://your-lambda-url.on.aws/xYz123"
}
```
The `expire_in_days` parameter is optional. If provided, the short URL will automatically expire after the specified number of days.

### Use a Short URL

Simply visit the short URL in a browser or make a GET request to it:

```bash
curl -L https://your-lambda-url.on.aws/xYz123
```

This will redirect to the original URL and increment the click count.

### Get URL Statistics

```bash
curl https://your-lambda-url.on.aws/stats/xYz123
```

Response:
```json
{
  "short_code": "xYz123",
  "original_url": "https://example.com/very/long/url/that/needs/shortening",
  "created_at": "2023-04-15T14:32:17Z",
  "expiration": "2023-04-22T14:32:17Z",
  "click_count": 42
}
```

## Customization

- Change the short code length by modifying the `codeLength` constant in the code
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

## Monitoring and Logging

- CloudWatch Logs are automatically configured for the Lambda function
- View logs in the AWS Console under CloudWatch Logs
- Consider setting up CloudWatch Alarms for error rates or high latency

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

## Future Enhancements

- Rate limiting: Prevent abuse of the service
- Authentication: Add user accounts to manage URLs
- Analytics dashboard: Visualize click data and trends
- Custom domain support: Use your own domain instead of the Lambda URL
- Custom shortcodes: Allow users to specify their own short codes
- Extended metadata: Store additional data like referrer or geolocation