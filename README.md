# SimpleBank

A production-grade banking backend service built in Go, exposing both a **gRPC API** and an **HTTP REST gateway**. Features include account management, fund transfers, transaction history, user authentication, and asynchronous email verification.

---

## Features

| Module | Description |
|---|---|
| Accounts | Create, read, list, and delete bank accounts with currency support |
| Entries | Track all debit/credit transactions per account |
| Transfers | Transfer funds between accounts with atomic balance updates |
| Users | User registration, login, and profile management |
| Sessions | PASETO-based session tokens with refresh support |
| Verify Email | Async email verification via background worker (Redis + Asynq) |
| RBAC | Role-Based Access Control on all API endpoints |

---

## Tech Stack

| Category | Technology |
|---|---|
| Language | Go 1.24 |
| HTTP Framework | Gin v1.11 |
| RPC Framework | gRPC + grpc-gateway v2 |
| Database | PostgreSQL 15 (pgx v5) |
| Query Generator | SQLC |
| Migrations | golang-migrate v4 |
| Auth | PASETO + JWT |
| Config | Viper v1.21 |
| Task Queue | Asynq + Redis 7 |
| Email | jordan-wright/email (Gmail SMTP) |
| Logging | zerolog v1.34 |
| Protobuf | protoc + openapiv2 |
| Validation | go-playground/validator |
| Testing | testify + mockgen (uber-go) |

---

## Prerequisites

Make sure the following tools are installed:

- [Go 1.24+](https://go.dev/dl)
- [Docker](https://docs.docker.com/get-docker)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [sqlc](https://sqlc.dev)
- [protoc](https://grpc.io/docs/protoc-installation)
- [mockgen](https://github.com/uber-go/mock) — `go install go.uber.org/mock/mockgen@latest`
- [evans](https://github.com/ktr0731/evans) — gRPC REPL client for testing
- [dbdocs](https://dbdocs.io) *(optional)* — `npm install -g dbdocs`

---

## Environment Configuration

Create an `app.env` file in the project root:

```env
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:3000
DB_SOURCE=postgres://root:<your_db_password>@localhost:5432/simple_bank?sslmode=disable
MIGRATION_URL=file://db/migration
HTTP_SERVER_ADDRESS=0.0.0.0:8080
GRPC_SERVER_ADDRESS=0.0.0.0:9090
TOKEN_SYMMETRIC_KEY=12345678901234567890123456789012
ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=24h
REDIS_ADDRESS=0.0.0.0:6379
EMAIL_SENDER_NAME=YourName
EMAIL_SENDER_ADDRESS=your_gmail@gmail.com
EMAIL_SENDER_PASSWORD=xxxx xxxx xxxx xxxx
```

> **Gmail App Password:** Go to [myaccount.google.com](https://myaccount.google.com) → Security → 2-Step Verification → App Passwords. Generate a password for "Mail" and paste it as `EMAIL_SENDER_PASSWORD`. **Never use your real Gmail password.**

---

## How to Run

### Step 1 — Create Docker Network

```bash
docker network create bank-network
```

### Step 2 — Start PostgreSQL

```bash
make postgres
```

Runs: `docker run --name postgres --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=your_password -d postgres`

### Step 3 — Start Redis

```bash
make redis
```

Runs: `docker run --name redis -p 6379:6379 -d redis:7-alpine`

### Step 4 — Create Database

```bash
make createdb
```

### Step 5 — Run Migrations

```bash
make migrateup        # Run all pending migrations
make migrateup1       # Run only the next 1 migration
```

### Step 6 — Start the Server

```bash
make server
```

This starts both the **gRPC server** on `:9090` and the **HTTP gateway** on `:8080`, plus the async email worker.

---

## All Make Commands

| Command | Description |
|---|---|
| `make postgres` | Start PostgreSQL container |
| `make redis` | Start Redis container |
| `make createdb` | Create the `simple_bank` database |
| `make dropdb` | Drop the database |
| `make migrateup` | Run all migrations |
| `make migrateup1` | Run next 1 migration |
| `make migratedown` | Rollback all migrations |
| `make migratedown1` | Rollback last 1 migration |
| `make new_migration name=<name>` | Create a new migration file |
| `make server` | Run the application |
| `make test` | Run all unit tests with coverage |
| `make sqlc` | Regenerate SQL query code |
| `make mock` | Regenerate mock files |
| `make db_schema` | Export DB schema SQL from DBML |
| `make protocc` | Regenerate protobuf Go files + Swagger |
| `make evans` | Open evans gRPC REPL client |

---

## API Usage

> **Currently active:** The app runs the **gRPC Gateway server** (not Gin). Both the HTTP gateway and gRPC server start together via `make server`.

### Switching Between Gin Server and Gateway Server

In `main.go`, there are two HTTP server implementations available. To switch, edit the `main()` function:

**Currently active — Gateway server (gRPC + HTTP/JSON gateway):**

```go
// main.go
runTaskProcessor(ctx, waitGroup, config, redisOpt, store)
runGateawayServer(ctx, waitGroup, config, store, taskDistributor)  // ← HTTP via grpc-gateway
runGrpcServer(ctx, waitGroup, config, store, taskDistributor)
```

**To switch to Gin server**, comment out `runGateawayServer` and call `runGinServer` instead:

```go
// main.go
runTaskProcessor(ctx, waitGroup, config, redisOpt, store)
// runGateawayServer(ctx, waitGroup, config, store, taskDistributor)
runGrpcServer(ctx, waitGroup, config, store, taskDistributor)

// add this after waitGroup.Wait():
runGinServer(config, store)
```

> **Note:** `runGinServer` is a blocking call (no graceful shutdown support), while `runGatewayServer` and `runGrpcServer` use `errgroup` with proper signal handling. Stick with the gateway server for production.

---

### gRPC API (Port 9090) + HTTP Gateway (Port 8080 `/v1/...`)

The gRPC server runs on port `9090`. It also exposes an HTTP/JSON gateway under the `/v1/` prefix via **grpc-gateway**.

**Authentication:** Protected endpoints require a Bearer token:

```
Authorization: Bearer <access_token>
```

**gRPC Methods & their HTTP gateway equivalents:**

| gRPC Method | HTTP Gateway Endpoint | Auth |
|---|---|---|
| `SimpleBank/CreateUser` | `POST /v1/create_user` | No |
| `SimpleBank/LoginUser` | `POST /v1/login_user` | No |
| `SimpleBank/UpdateUser` | `PATCH /v1/update_user` | Yes |
| `SimpleBank/VerifyEmail` | `GET /v1/verify_email?email_id=&secret_code=` | No |

**Example — Register User (HTTP gateway):**

```json
POST http://localhost:8080/v1/create_user
Content-Type: application/json

{
  "username": "john_doe",
  "full_name": "John Doe",
  "email": "john@example.com",
  "password": "secret123"
}
```

**Example — Login (HTTP gateway):**

```json
POST http://localhost:8080/v1/login_user
Content-Type: application/json

{
  "username": "john_doe",
  "password": "secret123"
}
```

Copy the `access_token` from the response and set it as a Bearer token in Postman's Authorization tab.

**Option A — evans CLI (gRPC directly):**

```bash
# Make sure the server is running first
make evans

# Inside evans REPL:
package pb
service SimpleBank
call CreateUser
```

**Option B — Postman gRPC:**

1. In Postman, click **New → gRPC Request**
2. Set server URL to `localhost:9090`
3. Click **Use Server Reflection** — Postman will auto-discover all RPC methods
4. Select a method (e.g. `SimpleBank/CreateUser`) and fill in the JSON body
5. For authenticated calls, go to the **Metadata** tab and add:

```
Key:   authorization
Value: Bearer <your_access_token>
```

---

### Gin HTTP API (Port 8080) — Inactive by Default

> Only available if you switch to `runGinServer` in `main.go` (see above).

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| POST | `/users` | Register new user | No |
| POST | `/users/login` | Login, get access + refresh token | No |
| POST | `/tokens/renew_access` | Renew access token | No |
| POST | `/accounts` | Create a bank account | Yes |
| GET | `/accounts/:id` | Get account by ID | Yes |
| GET | `/accounts` | List accounts (paginated) | Yes |
| PATCH | `/accounts/:id` | Update account | Yes |
| DELETE | `/accounts/:id` | Delete account | Yes |
| POST | `/transfers` | Transfer funds between accounts | Yes |
| GET | `/transfers/:id` | Get transfer by ID | Yes |
| GET | `/transfers` | List transfers | Yes |

---

## Email Verification

Email verification is handled asynchronously using **Asynq** (Redis-backed task queue).

**Flow:**

1. User registers via `POST /users`
2. Server enqueues a `SendVerifyEmail` task into Redis (via Asynq)
3. Background worker sends a verification email via Gmail SMTP
4. User clicks the link in the email — hits the gRPC gateway: `GET /v1/verify_email?email_id=<id>&secret_code=<code>`
5. Account is marked as verified in the database

The SMTP server used is `smtp.gmail.com:587`.

---

## Project Structure

```
.
├── api/                  # HTTP handlers (Gin) for REST gateway
├── db/
│   ├── migration/        # SQL migration files
│   ├── mock/             # Mock store for testing
│   └── sqlc/             # SQLC-generated type-safe queries
├── doc/                  # Swagger JSON, DBML schema, DB docs
├── gapi/                 # gRPC server handlers
├── mail/                 # Gmail SMTP email sender
├── pb/proto/             # Generated Go protobuf & gRPC bindings
├── proto/                # Protobuf definition files (.proto)
├── token/                # PASETO & JWT token implementations
├── util/                 # Config loader, random generators, password hashing
├── val/                  # Input validation helpers for gRPC
├── worker/               # Asynq task distributor and processor
└── .github/workflows/    # GitHub Actions CI/CD pipelines
```

---

## Quick Start

```bash
# 1. Create Docker network
docker network create bank-network

# 2. Start infrastructure
make postgres
make redis

# 3. Set up database
make createdb
make migrateup

# 4. Configure environment
cp app.env.example app.env   # fill in your values

# 5. Run the server
make server

# 6. (Optional) Run tests
make test

# 7. (Optional) Open gRPC REPL
make evans
```