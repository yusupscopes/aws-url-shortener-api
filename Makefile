.PHONY: build clean deploy

# Configuration
STACK_NAME ?= url-shortener
BINARY_NAME ?= bootstrap
REGION ?= ap-southeast-1
S3_BUCKET ?= $(STACK_NAME)-lambda-$(REGION)
AWS_PROFILE ?= ym3594216
# Build the Go binary for Lambda (Amazon Linux 2023)
build:
	@echo "Building Lambda function binary..."
	GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o $(BINARY_NAME) main.go
	chmod +x $(BINARY_NAME)
	zip function.zip $(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME) function.zip

# Create S3 bucket if it doesn't exist
create-bucket:
	@echo "Creating S3 bucket if it doesn't exist..."
	aws s3api head-bucket --bucket $(S3_BUCKET) 2>/dev/null --profile $(AWS_PROFILE) || aws s3 mb s3://$(S3_BUCKET) --region $(REGION) --profile $(AWS_PROFILE)

# Deploy the CloudFormation stack
deploy: build create-bucket
	@echo "Uploading Lambda function code to S3..."
	aws s3 cp function.zip s3://$(S3_BUCKET)/$(STACK_NAME)/function.zip --profile $(AWS_PROFILE)

	@echo "Deploying CloudFormation stack..."
	aws cloudformation deploy \
		--template-file template.yaml \
		--stack-name $(STACK_NAME) \
		--capabilities CAPABILITY_IAM \
		--parameter-overrides \
			S3Bucket=$(S3_BUCKET) \
			S3Key=$(STACK_NAME)/function.zip \
		--region $(REGION) \
		--profile $(AWS_PROFILE)

	@echo "Deployment complete."
	@echo "API URL:"
	@aws cloudformation describe-stacks \
		--stack-name $(STACK_NAME) \
		--query "Stacks[0].Outputs[?OutputKey=='UrlShortenerApiUrl'].OutputValue" \
		--output text \
		--region $(REGION) \
		--profile $(AWS_PROFILE)
