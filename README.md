# UniFriend API

## Overview

[Project overview to be provided by the user. Please describe what UniFriend does and the purpose of this API.]

## Table of Contents

- [Features](#features)
- [Technologies Used](#technologies-used)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
  - [Configuration](#configuration)
  - [Database Setup](#database-setup)
  - [Installation](#installation)
  - [Running the Application](#running-the-application)
- [API Endpoints](#api-endpoints)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Features

- User registration and authentication (JWT-based).
- Email verification for new user sign-ups.
- User profile management, including profile picture uploads (to AWS S3) and additional image galleries.
- Quiz system:
    - Fetching quiz questions.
    - Saving user answers.
    - Retrieving quiz results for users.
- Management of academic majors or areas of interest.
- Secure file uploads to AWS S3.
- Email sending capabilities via AWS SES.
- Comprehensive API documentation via Swagger.
- Health check endpoint for monitoring.

## Technologies Used

- **Backend:** Go (Golang)
- **Web Framework:** Gin
- **ORM:** GORM
- **Database:** MySQL (with SQLite for testing)
- **Authentication:** JWT (JSON Web Tokens)
- **Cloud Services:**
    - AWS S3 (Simple Storage Service) for file storage.
    - AWS SES (Simple Email Service) for email sending.
- **API Documentation:** Swagger / OpenAPI
- **Containerization:** Docker
- **Continuous Integration:** GitHub Actions

## Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/doc/install): Version 1.22.0 or higher (as per `go.mod` toolchain)
- [MySQL](https://dev.mysql.com/downloads/): For database.
- [Docker](https://www.docker.com/get-started): For containerized deployment and development (optional but recommended).
- [Git](https://git-scm.com/): For version control.
- An AWS account with configured credentials if you intend to use S3 and SES features.

## Getting Started

Follow these instructions to get a copy of the project up and running on your local machine for development and testing purposes.

### Configuration

1.  **Environment Variables:**
    The project uses environment variables for configuration. These are stored in a `.env` file.

    *   Copy the example environment file:
        ```bash
        cp .env.example .env
        ```
    *   Open the `.env` file and fill in the required values. Below is a description of each variable:

        | Variable                          | Description                                                                 | Example        |
        | --------------------------------- | --------------------------------------------------------------------------- | -------------- |
        | `DB_HOST`                         | Hostname or IP address of your MySQL database server.                       | `localhost`    |
        | `DB_DRIVER`                       | Database driver to use (should be `mysql`).                                 | `mysql`        |
        | `DB_USER`                         | Username for connecting to the MySQL database.                              | `root`         |
        | `DB_PASSWORD`                     | Password for the MySQL database user.                                       | `yourpassword` |
        | `DB_NAME`                         | Name of the MySQL database to use.                                          | `unifriend_db` |
        | `DB_PORT`                         | Port number for the MySQL database server.                                  | `3306`         |
        | `TOKEN_HOUR_LIFESPAN`             | Lifespan of JWT authentication tokens in hours.                             | `24`           |
        | `API_SECRET`                      | Secret key used for signing JWT tokens.                                     | `your-secret-key` |
        | `AWS_ACCESS_KEY_ID`               | Your AWS Access Key ID (for S3 and SES).                                    |                |
        | `AWS_SECRET_KEY`                  | Your AWS Secret Access Key (for S3 and SES).                                |                |
        | `AWS_SESSION_TOKEN`               | Your AWS Session Token (if using temporary credentials).                    |                |
        | `AWS_REGION`                      | The AWS region where your S3 bucket and SES are configured.                 | `us-east-1`    |
        | `AWS_BUCKET_NAME`                 | Name of the AWS S3 bucket used for file storage.                            | `your-bucket-name` |
        | `AWS_EMAIL`                       | Email address registered with AWS SES, used as the sender for verification emails. | `noreply@example.com` |
        | `MAX_SIZE_PROFILE_IMAGE_KB`       | Maximum allowed size for profile image uploads in kilobytes.                | `1024`         |
        | `CLIENT_DOMAIN`                   | The domain of the client application (used for CORS configuration).         | `http://localhost:3000` |
        | `VERIFICATION_CODE_LIFESPAN_MINUTES` | Lifespan of email verification codes in minutes.                          | `15`           |

    **Note:** For AWS credentials, it's recommended to use IAM roles or configure them via the AWS CLI shared credentials file (`~/.aws/credentials`) if running locally, rather than hardcoding them directly in the `.env` file for production environments. The application will pick them up from the environment if `AWS_ACCESS_KEY_ID` and `AWS_SECRET_KEY` are not set in the `.env` file.

### Database Setup

1.  **Ensure MySQL is running:** Make sure you have a MySQL server instance running and accessible.
2.  **Create the database:** Connect to your MySQL server and create the database that you specified in the `DB_NAME` environment variable.
    ```sql
    CREATE DATABASE your_database_name; -- Replace your_database_name with the actual name
    ```
3.  **Database Migrations:**
    The application uses GORM for database interactions. The database schema is defined by the structs in the `models/` directory.
    For test environments, migrations are run automatically using an in-memory SQLite database (see `models/setup.go` - `SetupTestDB()`).
    For development and production environments using MySQL, the `AutoMigrate` function in `models/setup.go` (`ConnectDataBase()`) is currently commented out. You will need to either:
    *   Uncomment the `DB.AutoMigrate(...)` lines in `models/setup.go` temporarily to run migrations when the application starts. Remember to comment it back if you don't want it running on every startup.
    *   Or, manage database migrations manually using a migration tool of your choice, ensuring the schema matches the model definitions. The models to be migrated are: `User`, `Major`, `OptionTable`, `QuestionTable`, `QuizTable`, `UserResponse`, `EmailDomains`, `EmailsVerification`, and `UsersImages`.

### Installation

1.  **Clone the repository:**
    ```bash
    git clone <your-repository-url> # Replace <your-repository-url> with the actual URL
    cd unifriend-api
    ```
2.  **Install Go dependencies:**
    Download the necessary Go modules:
    ```bash
    go mod tidy
    ```

### Running the Application

You can run the application in two main ways:

**1. Using Go directly (with live reload via `air`):**

This method is useful for active development as it provides live reloading of the application when code changes are detected. The `Dockerfile` installs `air`, a live reloading tool for Go applications. Ensure you have `air` installed locally if you prefer not to use Docker for development. If you don't have `air` installed locally, you can install it by running: `go install github.com/air-verse/air@latest`.

*   **Ensure all dependencies are installed:**
    ```bash
    go mod tidy
    ```
*   **Run the application using `air` (recommended for development):**
    This will watch for file changes and automatically rebuild and restart the application. The configuration for `air` is in `.air.toml`.
    ```bash
    air -c .air.toml
    ```
*   **Run the application using standard `go run` (without live reload):**
    ```bash
    go run main.go
    ```
The API will be accessible at `http://localhost:8090`.

**2. Using Docker Compose (recommended for consistency):**

This method uses Docker to build and run the application in a containerized environment, as defined in `Dockerfile` and `docker-compose.yml`. This is often simpler for ensuring a consistent environment.

*   **Build and start the services:**
    From the root of the project, run:
    ```bash
    docker-compose up --build
    ```
    To run in detached mode (in the background):
    ```bash
    docker-compose up --build -d
    ```
*   **Accessing the application:**
    The API will be accessible at `http://localhost:8090`.
*   **Stopping the services:**
    ```bash
    docker-compose down
    ```

Choose the method that best suits your development workflow.
