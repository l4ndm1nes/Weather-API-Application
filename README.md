# Weather API Application

Weather API Application allows users to subscribe to weather updates for their city. The API provides functionality for subscribing, confirming, unsubscribing, and retrieving weather data.

### Deployed Project

The project is deployed and available at [https://weather-api-app-csei.onrender.com/](https://weather-api-app-csei.onrender.com/).

## Features

- **Weather Forecast**: Get the current weather for any city.
- **Subscription Management**: Subscribe to receive weather updates at regular intervals (hourly or daily).
- **Email Confirmation**: Users must confirm their subscription via email.
- **Unsubscribe**: Unsubscribe from weather updates using the provided token.

## API Endpoints

### 1. `/weather`
- **Method**: `GET`
- **Description**: Retrieve the current weather for a city.
- **Query Parameters**:
    - `city`: City name (Latin letters only) for weather forecast (Required)
- **Responses**:
    - `200 OK`: Successful weather retrieval
    - `400 Bad Request`: Invalid input
    - `404 Not Found`: City not found

### 2. `/subscribe`
- **Method**: `POST`
- **Description**: Subscribe to weather updates.
- **Form Parameters**:
    - `email`: User's email address (Required)
    - `city`: City for weather updates (Required)
    - `frequency`: Frequency of updates (`hourly` or `daily`) (Required)
- **Responses**:
    - `200 OK`: Subscription successful. Confirmation email sent.
    - `400 Bad Request`: Invalid input
    - `409 Conflict`: Email already subscribed

### 3. `/confirm/{token}`
- **Method**: `GET`
- **Description**: Confirm email subscription using the confirmation token sent in the email.
- **Path Parameters**:
    - `token`: Confirmation token (Required)
- **Responses**:
    - `200 OK`: Subscription confirmed successfully
    - `400 Bad Request`: Invalid token
    - `404 Not Found`: Token not found

### 4. `/unsubscribe/{token}`
- **Method**: `GET`
- **Description**: Unsubscribe from weather updates using the unsubscribe token sent in the email.
- **Path Parameters**:
    - `token`: Unsubscribe token (Required)
- **Responses**:
    - `200 OK`: Unsubscribed successfully
    - `400 Bad Request`: Invalid token
    - `404 Not Found`: Token not found

## Swagger Documentation

The API documentation can be accessed through Swagger, which is available at the following URL after deployment:

Swagger URL: https://weather-api-app-csei.onrender.com/swagger

Swagger provides a complete visual representation of the API, where you can explore and test the available endpoints interactively.

## Setup

### Prerequisites
- Docker
- Docker Compose
- Makefile (optional for easier management)

### Running Locally

1. Clone the repository:
```bash
git clone https://github.com/l4ndm1nes/Weather-API-Application.git
```

2. Navigate to the project directory:

```bash
cd Weather-API-Application
```

### Create a .env file for local environment configuration (use .env.example as a reference):

```bash
cp .env.example .env
```

### Add the required values in .env:

- **DB_HOST**: Database host (for local development, use localhost)
- **DB_PORT**: Database port (default: 5432)
- **DB_USER**: Database username (default: postgres)
- **DB_PASSWORD**: Database password
- **DB_NAME**: Database name (default: weather)
- **DB_SSLMODE**: Database SSL mode (for local: disable, for production: require)
- **SMTP_HOST**: SMTP mail host (e.g., smtp.mailtrap.io)
- **SMTP_PORT**: SMTP port (e.g., 2525)
- **SMTP_USER**: SMTP username
- **SMTP_PASS**: SMTP password
- **SMTP_FROM**: From email address
- **WEATHER_API_KEY**: API key for weather data
- **BASE_URL**: The base URL of your app (for local: http://localhost:8080, for production: your deployed URL)

### Build and run the project using Docker:

```bash
docker-compose up --build
```

Open `http://localhost:8080` in your browser to test the API.

---

## Using Makefile

The project includes a Makefile to simplify common tasks like building, testing, and running migrations. Here are the available commands:

### Migrate Script Ready: Prepares the migration script with execute permissions.

```bash
make migrate-script-ready
```

### Migrate Up: Applies all pending migrations.

```bash
make migrate-up
```

### Migrate Down: Rolls back the last applied migration.

```bash
make migrate-down
```

### Run: Build and run the Docker containers.

```bash
make run
```

### Build: Build the Go application.

```bash
make build
```

### Fmt: Format the Go code.

```bash
make fmt
```

### Lint: Run linting on the Go code using golangci-lint.

```bash
make lint
```

### Test: Run tests for the application.

```bash
make test
```

---

## Deploy to Render

To deploy this app to Render:

1. Push your repository to GitHub.
2. Go to Render.
3. Create a new Web Service.
4. Connect your GitHub repository.
5. Set up environment variables in the Render dashboard.
6. Click "Deploy" to start the deployment.

---

## Testing

Run tests using the following command:

```bash
make test
```

You can also test the endpoints using curl or Postman:

### GET weather:

```bash
curl -X GET "http://localhost:8080/api/weather?city=Kyiv"
```

### POST subscribe:

```bash
curl -X POST "http://localhost:8080/api/subscribe" -d "email=user@example.com&city=Kyiv&frequency=daily"
```

### GET confirm:

```bash
curl -X GET "http://localhost:8080/api/confirm/{token}"
```

### GET unsubcribe:

```bash
curl -X GET "http://localhost:8080/api/unsubsribe/{token}"
```

## Project Structure
```bash
.
Weather-API-Application/
│
├── cmd/
│   └── httpserver/                    - Main server entry point
│
├── docs/                              - Swagger documentation
│
├── internal/                          - Core application logic
│   ├── adapter/                       - Adapter for interacting with repositories, external services, and APIs
│   ├── config/                        - Application configuration and settings
│   ├── handler/                       - API request handlers
│   ├── mocks/                         - Mock objects for testing
│   ├── model/                         - Data models and structures
│   ├── scheduler/                     - Periodic job tasks and scheduling
│   └── service/                       - Business logic services
│
├── migrations/                        - Database migrations
│
├── pkg/                               - Utility function packages
│   ├── middleware/                    - Middleware for request processing
│   
├── scripts/                           - Scripts for operations, such as waiting for the database
│
├── test/                              - Tests for verifying application functionality
│   ├── integration/                   - Integration tests
│   └── unit/                          - Unit tests
│
├── web/                               - Files for the web interface
│
├── Dockerfile                         - Dockerfile for building the containerized application
│
├── docker-compose.yml                 - Docker Compose configuration for setting up the app with services
│
├── Makefile                           - Makefile for automation of tasks
│
├── README.md                          - Project documentation and setup instructions
```

