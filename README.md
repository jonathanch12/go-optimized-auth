# Optimized Go Auth API

A small JWT authentication API built with [Gin](https://github.com/gin-gonic/gin). It uses a **hybrid JWT + Redis session model**: the JWT proves the token was issued by this server, and Redis is the server-side source of truth for whether that session is still active. This gives the API JWT's stateless request format plus server-side control for immediate logout and token revocation.

User data is held in an in-memory store (no database), seeded with a single user at startup. Active sessions live in Redis, keyed by the JWT `jti`, with a TTL that matches the token's lifetime so expired sessions clean themselves up.

## Features

- `POST /api/auth/login` — validate credentials, issue a signed JWT, and record an active session in Redis.
- `GET /api/auth/me` — return the authenticated user's profile (protected; validates the JWT **and** the Redis session).
- `POST /api/auth/logout` — revoke the current session by deleting it from Redis (immediate, server-side logout).
- JWTs carry a unique `jti` (UUID) used as the Redis session key.
- Passwords hashed with bcrypt; tokens signed with HS256 (golang-jwt v5), with signing-algorithm enforcement on parse.
- Sessions stored in Redis via a `SessionStore` abstraction (swappable for memory/SQL later) with automatic TTL expiry.
- Fails closed: if Redis is unreachable, protected requests are rejected rather than allowed through.
- Config via `.env` (JWT secret, token expiry, port, Redis connection).

## How it works

| Concern                        | Handled by |
| ------------------------------ | ---------- |
| Cryptographic proof the token was issued by this server | JWT signature + `exp` validation |
| Whether the issued session is still active | Redis key `auth:session:<jti>` |

On **login**, after credentials are verified, the server mints a JWT with a random `jti` and runs `SET auth:session:<jti> <user_id> EX <ttl>`. On every **protected request**, the middleware validates the JWT, extracts the `jti`, and checks Redis — a missing key means the session was revoked or expired, so the request is rejected. On **logout**, the server runs `DEL auth:session:<jti>`; the client may still hold the JWT, but the server no longer recognizes the session.

## Project structure

```
go-auth-api/
├── cmd/
│   └── api/
│       └── main.go              # Entry point: loads env, connects+pings Redis, wires everything, starts the server
├── internal/
│   ├── auth/
│   │   ├── jwt.go               # GenerateToken (adds jti/exp/iat, returns token+jti+ttl) / ParseToken (HS256, jti+user_id required)
│   │   └── password.go          # HashPassword / CheckPassword (bcrypt)
│   ├── handler/
│   │   └── auth_handler.go      # Login, Me, Logout HTTP handlers
│   ├── middleware/
│   │   └── auth.go              # RequireAuth(sessionStore): validates JWT + Redis session, sets userID/jti on context
│   ├── model/
│   │   ├── user.go              # User struct (password excluded from JSON)
│   │   └── access_token.go      # AccessToken domain model (persistence-independent, ready for future SQL)
│   ├── router/
│   │   └── router.go            # Route registration under /api/auth, middleware applied per-route
│   └── store/
│       ├── user_store.go        # In-memory user store, seeded with one user
│       └── session_store.go     # SessionStore interface + RedisSessionStore (auth:session:<jti> with TTL)
├── docker-compose.yml           # Redis service for local development
├── .env                         # Local config (not committed)
├── .env.example                 # Template for .env
├── go.mod
└── go.sum
```

## Requirements

- Go 1.27+ (see `go.mod`)
- [Docker](https://www.docker.com/) (to run Redis locally)
- [Postman](https://www.postman.com/) for testing the routes

## Setup

1. Clone the repository and move into the project folder.

2. Create your `.env` from the template:

   ```bash
   cp .env.example .env
   ```

   `.env` holds:

   | Variable              | Description                                      | Example                              |
   | --------------------- | ------------------------------------------------ | ------------------------------------ |
   | `JWT_SECRET`          | Secret used to sign and verify tokens (HS256)    | `change-me-to-a-long-random-string`  |
   | `JWT_EXPIRES_MINUTES` | Token lifetime in minutes (also the Redis TTL)   | `60`                                 |
   | `PORT`                | Port the server listens on                       | `8080`                               |
   | `REDIS_ADDR`          | Redis host:port                                  | `localhost:6379`                     |
   | `REDIS_PASSWORD`      | Redis password (leave empty if none)             | ``                                   |
   | `REDIS_DB`            | Redis logical database number                    | `0`                                  |

   Use a long, random value for `JWT_SECRET`. The `.env` file is git-ignored so the secret is not committed.

3. Start Redis with Docker:

   ```bash
   docker compose up -d
   ```

   This starts a `redis:7-alpine` container on port `6379` with append-only persistence and a healthcheck. Stop it later with `docker compose down` (add `-v` to also remove the data volume).

4. Install dependencies:

   ```bash
   go mod download
   ```

## Running the server

Make sure Redis is running first (`docker compose up -d`), then from the project root:

```bash
go run ./cmd/api
```

On startup the server connects to Redis and **pings it**; if Redis is unreachable it exits with a fatal error rather than starting in a broken state. Once connected, Gin prints the registered routes and the listening address (default port `8080`). Stop the server with `Ctrl+C`.

### Seeded user

The in-memory store is seeded with one user you can use to log in:

| Field    | Value       |
| -------- | ----------- |
| username | `test1`     |
| password | `test123#`  |

## Testing in Postman

> Postman collection / docs: [Postman Collection](https://documenter.getpostman.com/view/57333016/2sBYHPz25e)

Make sure both Redis and the server are running before sending requests. Base URL: `http://localhost:8080`.

### 1. Login — get a token

- **Method:** `POST`
- **URL:** `http://localhost:8080/api/auth/login`
- **Body:** select **raw** → **JSON**, then:

  ```json
  {
    "username": "test1",
    "password": "test123#"
  }
  ```

- **Expected response:** `200 OK`

  ```json
  {
    "token": "<jwt-token-string>"
  }
  ```

Copy the `token` value — you'll need it for the next calls. Behind the scenes, an `auth:session:<jti>` key is now set in Redis.

### 2. Me — get the current user (valid token)

- **Method:** `GET`
- **URL:** `http://localhost:8080/api/auth/me`
- **Authorization:** select type **Bearer Token** and paste the token from step 1 (paste only the token, not the word `Bearer` — Postman adds it).
- **Expected response:** `200 OK`

  ```json
  {
    "id": 1,
    "name": "Test",
    "username": "test1"
  }
  ```

### 3. Logout — revoke the session

- **Method:** `POST`
- **URL:** `http://localhost:8080/api/auth/logout`
- **Authorization:** Bearer Token set to the token from step 1.
- **Expected response:** `204 No Content` (empty body).

  This deletes the session from Redis. Logout is idempotent: calling it again (or after the token's TTL has expired) still returns `204`, because the end state — "session not active" — is already true. Only a Redis connectivity failure returns `500`.

### 4. Me — after logout (revoked token)

- **Method:** `GET`
- **URL:** `http://localhost:8080/api/auth/me`
- **Authorization:** Bearer Token set to the **same** token from step 1.
- **Expected response:** `401 Unauthorized`

  ```json
  {
    "error": "Unauthorized"
  }
  ```

  The JWT is still cryptographically valid and unexpired, but its Redis session no longer exists — demonstrating immediate server-side revocation.

### 5. Me — no token

- **Method:** `GET`
- **URL:** `http://localhost:8080/api/auth/me`
- **Authorization:** none.
- **Expected response:** `401 Unauthorized`

  ```json
  {
    "error": "Unauthorized"
  }
  ```

### 6. Me — bogus token

- **Method:** `GET`
- **URL:** `http://localhost:8080/api/auth/me`
- **Authorization:** Bearer Token set to a garbage value like `abc.def.ghi`.
- **Expected response:** `401 Unauthorized`

  ```json
  {
    "error": "Unauthorized"
  }
  ```

### 7. Login — wrong credentials

- **Method:** `POST`
- **URL:** `http://localhost:8080/api/auth/login`
- **Body:** raw → JSON with a wrong password:

  ```json
  {
    "username": "test1",
    "password": "wrong-password"
  }
  ```

- **Expected response:** `401 Unauthorized`

  ```json
  {
    "error": "Invalid credentials"
  }
  ```

  The same `Invalid credentials` response is returned whether the username is unknown or the password is wrong, so the API does not reveal which field was incorrect.

## API reference

| Method | Path                 | Auth required | Description                                        |
| ------ | -------------------- | ------------- | -------------------------------------------------- |
| POST   | `/api/auth/login`    | No            | Validate credentials, issue a JWT, create session  |
| GET    | `/api/auth/me`       | Yes (Bearer)  | Return the current user's profile                   |
| POST   | `/api/auth/logout`   | Yes (Bearer)  | Revoke the current session (204 No Content)         |

## Notes on persistence

The `AccessToken` model (`internal/model/access_token.go`) is deliberately independent of its storage. Today, active sessions live only in Redis. If a database is introduced later, the same model maps to an `access_tokens` table for persistent/audit history, while Redis remains the fast runtime session store — without changing the authentication flow.
