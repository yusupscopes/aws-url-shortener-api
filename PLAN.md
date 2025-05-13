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


## Phase 1: Project Setup & Repo Structure

- [x] **Initialize Project:**
  - Set up Go module: `go mod init github.com/yourusername/url-shortener`
  - Create a repository with the following folder structure:
    ```
    url-shortener/
      ├── cmd/
      │   └── main.go             # Lambda entrypoint
      ├── pkg/
      │   ├── database/           # DynamoDB integration
      │   ├── handler/            # Request handlers
      │   ├── model/              # Data models
      │   └── utils/              # Utility functions
      ├── template.yaml           # AWS CloudFormation template for infrastructure
      ├── go.mod                  # Go module file
      ├── go.sum                  # Go sum file
      ├── README.md               # Project documentation
      └── Makefile                # Build & deploy commands (SAM based)
    ```

- [x] **Set Up Basic Code:**
  - Create a basic Lambda handler in `cmd/main.go`.
  - Add placeholder logic in `pkg/handler.go` for URL shortening and redirection.
  - Define data models in `pkg/model.go` and helper functions in `pkg/utils.go`.

---

## Phase 2: Implement URL Shortening (POST /shorten)

- [x] **Lambda Handler:**
  - Implement the endpoint to generate a unique short code using a random base62 string or a hash.
  - Validate the user-provided URL input.

- [x] **DynamoDB Integration:**
  - Store new URL mappings with details:
    - `shortCode` (primary key)
    - `originalURL`
    - `createdAt` timestamp
    - `expiration` timestamp (optional)
    - `clickCount` initialized to zero

- [x] **Response:**
  - Return a JSON response containing the full shortened URL using the Lambda Function URL.

---

## Phase 3: Implement Redirection (GET /{shortCode})

- [x] **Lookup and Redirect:**
  - Retrieve the original URL from DynamoDB using the short code.
  - Perform an atomic update to increment the click count.
  - Return an HTTP 301/302 redirect with the `Location` header set to the original URL.

---

## Phase 4: Optional Analytics Endpoint (GET /stats/{shortCode})

- [x] **Analytics Data:**
  - Create an endpoint to retrieve analytics for a given short URL:
    - Original URL
    - Creation date
    - Expiration date (if applicable)
    - Total click count

- [x] **Response:**
  - Return a formatted JSON object with the analytics data.

---

## Phase 5: Expiration Support

- [x] **TTL Configuration:**
  - Add an expiration attribute to new URL mappings.
  - Enable and configure DynamoDB’s Time-to-Live (TTL) on the `expiration` attribute to automatically delete expired entries.

---

## Phase 6: Logging, Testing, and Monitoring

- [x] **Logging:**
  - Integrate AWS CloudWatch Logs for debugging and monitoring.
  
- [ ] **Testing:**
  - Write unit tests covering URL generation, DynamoDB operations, and request handling logic.
  - Create integration tests using tools like Postman or curl.

- [ ] **Monitoring:**
  - Set up CloudWatch metrics and alarms if needed.

---

## Phase 7: Documentation & Deployment

- [ ] **Documentation:**
  - Update `README.md` with detailed usage instructions, the architecture overview, and deployment steps (using SAM).
  
- [ ] **Infrastructure Deployment:**
  - Use AWS SAM to build and deploy:
    - Lambda function with a Function URL
    - DynamoDB table with TTL enabled
  - Validate endpoint functionality post-deployment.

- [ ] **Deployment Automation:**
  - Utilize the provided `Makefile` to streamline build and deployment steps.

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

## ✅ Future Enhancements

- **Custom Shortcodes:** Allow users to specify a custom shortcode instead of a randomly generated one.
- **Authentication:** Implement authentication for a private API or to restrict modification/deletion of URL mappings.
- **Rate Limiting:** Enforce usage quotas or rate limits to manage abuse and control traffic.
- **Analytics Dashboard:** Create a web-based dashboard to visualize click analytics and trends.
- **Custom Domain Support:** Enable the use of custom domains instead of the default Lambda Function URL.
- **Extended Metadata:** Add support for storing additional metadata (e.g., referrer, geolocation data).

---

## ✅ Summary

This PLAN document details the steps to build a robust, serverless URL Shortener service using AWS Lambda Function URLs and DynamoDB, along with future paths for enhancement and refinement. By following this phased approach, developers can systematically build, test, and deploy a production-ready cloud-native microservice.