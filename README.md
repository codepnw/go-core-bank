# 🏦 Core Banking / E-Wallet API

A robust, high-performance backend API for processing financial transactions. Built with **Go** and **PostgreSQL**, this project strongly focuses on database ACID properties, deadlock prevention, and Clean Architecture principles.

## 🚀 Key Features

- **ACID Transactions:** Ensures all money transfers are processed reliably with an "all-or-nothing" approach. No funds are lost or duplicated during unexpected system or network failures.
- **Deadlock Prevention:** Mitigates database race conditions and deadlocks using **Consistent Lock Ordering** (ordering account IDs before acquiring row-level locks).
- **Clean Architecture:** Strictly separates concerns across HTTP Handlers, Business Logic (Services), and Data Access (Repositories) to ensure maintainability, scalability, and high testability.
- **Fail-Fast Validation:** Validates core business rules directly at the Service layer (e.g., self-transfer prevention, insufficient funds checks, zero-amount transfers) to minimize unnecessary database hits and optimize performance.
- **Pragmatic Testing Strategy:** Focuses testing purely on the core transfer logic. Includes **Unit Tests** (using mocks for business rules) and highly concurrent **Integration Tests** (using Go's Goroutines and Channels to verify database row-level locks under heavy load).

## 🗄️ Domain Models

The system is built around three core entities:
1. **Account:** Stores user account details and the current balance.
2. **Entry:** Records all individual balance changes (deposits/withdrawals) for precise audit trails.
3. **Transfer:** Records complete transfer transactions between two accounts (From -> To).


## 🏗️ Project Structure & Initialization

> **💡 Note on Commit History:** This repository was initialized using a standard **[Go Starter Kit](https://github.com/codepnw/go-starter-kit)** to handle repetitive boilerplate code (e.g., basic folder structure, router setup). This strategic choice allowed the development focus to remain entirely on building the complex core business logic, database transactions, and concurrency handling that you will see in the subsequent commits.

The project directory follows a modular Go layout, combining **Clean Architecture** principles with a **Package by Feature** structure for better maintainability and high cohesion:


```text
.
├── cmd
│   └── api                 # Application entry point (main.go)
├── internal
│   ├── auth                # Context management & user session extraction
│   ├── config              # Environment variables & configuration loader
│   ├── errs                # Centralized custom error types
│   ├── features            # Domain modules (User, Account, Transfer) containing their own Handler, Service, and Repo
│   ├── middleware          # HTTP middlewares (e.g., JWT Auth, Logger, Recovery)
│   └── server              # HTTP server initialization & route registration
├── pkg
│   ├── database            # Database connection & transaction management
│   ├── jwttoken            # JWT generation & validation utilities
│   └── utils               # Shared helper functions (e.g., password hashing)
├── .air.toml               # Live reload configuration for rapid local development
├── .env.example            # Template for environment variables
├── docker-compose.yml      # Local development environment setup (PostgreSQL, etc.)
├── Dockerfile              # Docker build instructions for production deployment
└── Makefile                # Shortcut commands for build, test, migrate, and run
```


## 🚀 Getting Started

Follow these steps to get the project up and running on your local machine.

### Option 1: Quick Start with Docker (Recommended) 🐳

This will spin up both the Go API server and the PostgreSQL database container.

1.  **Clone the repository**
    ```bash
    git clone https://github.com/codepnw/go-core-bank.git

    cd go-core-bank
    ```

2.  **Setup Environment Variables**
    ```bash
    cp -n .env.example .env
    ```
    *Modify `.env` if you want to change default ports or secrets.*

3.  **Start Services**
    ```bash
    # Build and start both App & DB containers
    docker compose up -d --build
    ```
    *(Note: If you only want to run the database container, use `docker compose up -d db`)*

4.  **Run Database Migrations**
    ```bash
    # If you have Makefile configured
    make migrate-up
    
    # Or manually using golang-migrate
    migrate -path [MIGRATION_PATH] -database [DATABASE_URL] up
    ```

The API will be available at `http://localhost:8080/api/v1` (Default URL).

### Option 2: Run Locally (Without Docker)

If you prefer to run the Go application directly on your host machine:

1.  **Start PostgreSQL** (Make sure you have a running instance).
2.  **Update `.env`** to point to your local DB credentials.
3.  **Run the application**:
    ```bash
    # If you have Makefile configured
    make run

    # If you have Air Live Reload
    air

    # Or run the standard Go command
    go run cmd/api/main.go

    ```

---

## 📡 API Endpoints Summary

The API is grouped into logical domains. All protected routes require a valid JWT Bearer Token in the `Authorization` header.

*(Base URL: `http://localhost:8080/api/v1`)*

### 🔐 Authentication & Users
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `POST` | `/auth/register` | Register a new user | ❌ |
| `POST` | `/auth/login` | Authenticate and receive JWT tokens | ❌ |
| `POST` | `/auth/refresh` | Refresh an expired access token | ❌ |
| `POST` | `/auth/logout` | Invalidate user session | 🔒 Yes |
| `GET`  | `/users/profile` | Get current logged-in user details | 🔒 Yes |

### 🏦 Accounts Management
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `POST` | `/accounts/` | Create a new bank account | 🔒 Yes |
| `GET`  | `/accounts/` | List all accounts (with pagination) | 🔒 Yes |
| `GET`  | `/accounts/:id` | Get account details and current balance | 🔒 Yes |
| `POST` | `/accounts/:id/deposit` | Deposit money into an account | 🔒 Yes |
| `POST` | `/accounts/:id/withdraw`| Withdraw money from an account | 🔒 Yes |

### 💸 Transfers (Core Business Logic)
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `POST` | `/transfers/` | Transfer money between two accounts | 🔒 Yes |

### 🩺 System
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `GET`  | `/health` | Check API server status | ❌ |
