# URL Shortener API (Go + AWS Lambda Function URL)

Create and manage short URLs with optional expiration and click analytics using Go, AWS Lambda Function URLs, and DynamoDB.

---

## ✅ Tech Stack

- **Language**: Go (Golang)
- **Compute**: AWS Lambda (with Function URL)
- **Database**: AWS DynamoDB (with TTL & analytics support)
- **Infra-as-Code**: AWS CloudFormation
- **Logging**: AWS CloudWatch Logs

---

## ✅ Functional Features

- `POST /shorten` — create a new short URL
- `GET /{shortCode}` — redirect to original URL and increment click count
- `GET /stats/{shortCode}` — (optional) return analytics

---

## ✅ Phase-Based Plan

### Phase 1: Project Setup & Repo Structure

- [ ] Initialize Go module: `go mod init`
- [ ] Define basic folder structure:
```
/cmd/main.go
/pkg/
handler.go
model.go
utils.go
template.yaml (for SAM) or main.tf (for Terraform)
```
- [ ] Set up simple build + deploy flow using AWS CLI or IaC tool

---

### Phase 2: Implement URL Shortening (POST /shorten)

- [ ] Create URL shortening Lambda handler
- [ ] Generate unique short codes (random base62 or hash)
- [ ] Save mapping to DynamoDB:
- shortCode (PK)
- originalURL
- createdAt
- expiration (optional)
- clickCount = 0
- [ ] Return shortened URL using Lambda Function URL

---

### Phase 3: Implement Redirection (GET /{shortCode})

- [ ] Retrieve original URL from DynamoDB
- [ ] Update click count (atomic increment)
- [ ] Return HTTP 301/302 redirect response

---

### Phase 4: Optional Analytics (GET /stats/{shortCode})

- [ ] Add endpoint to return:
- originalURL
- createdAt
- expiration
- total clicks
- [ ] Format and return JSON

---

### Phase 5: Expiration Support

- [ ] Add optional TTL to each entry
- [ ] Configure DynamoDB TTL attribute
- [ ] Clean up expired entries automatically

---

### Phase 6: Logging, Testing, and Monitoring

- [ ] Add CloudWatch logs for each operation
- [ ] Write unit tests for logic functions
- [ ] Add simple integration test (e.g., curl or Postman collection)

---

### Phase 7: Documentation & Deployment

- [ ] Write detailed README with:
- Project description
- Architecture diagram
- Deployment instructions
- API usage examples
- [ ] Publish live demo URL (Lambda Function URL)
- [ ] (Optional) Write a blog or portfolio post

---

## ✅ DynamoDB Table Design

| Attribute     | Type      | Description                          |
|---------------|-----------|--------------------------------------|
| shortCode     | String (PK) | Unique code used in shortened URL    |
| originalURL   | String    | Full original long URL               |
| createdAt     | String    | ISO timestamp                        |
| expiration    | Number    | Unix timestamp for TTL               |
| clickCount    | Number    | Total number of redirects            |

---

## ✅ Example Shortened URL

```
POST https://<lambda-function-url>/shorten
{
    "url": "https://example.com/very/long/link",
    "expire_in_days": 7
}

Response:
{
    "short_url": "https://<lambda-url>/xYz123"
}
```
---

## ✅ Notes

- Avoid API Gateway to reduce cost and simplify setup.
- Lambda Function URLs support basic HTTP handling.
- Consider IAM or custom headers if you want private access or rate limiting.

