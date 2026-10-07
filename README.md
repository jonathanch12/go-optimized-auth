# go-auth-api

A small JWT authentication API built with [Gin](https://github.com/gin-gonic/gin). It exposes a login endpoint that issues a signed JWT and a protected endpoint that returns the current user's profile. User data is held in an in-memory store (no database), seeded with a single user at startup.

## Features

- `POST /api/auth/login` — validate credentials and return a JWT.
- `GET /api/auth/me` — return the authenticated user's profile (protected by Bearer-token middleware).
- Passwords hashed with bcrypt; tokens signed with HS256 (golang-jwt v5).
- Config via `.env` (JWT secret, token expiry, port).

## Project structure

```
go-auth-api/
├── cmd/
│   └── api/
│       └── main.go          # Entry point: loads env, wires store/handler/router, starts the server
├── internal/
│   ├── auth/
│   │   ├── jwt.go           # GenerateToken / ParseToken (HS256, alg-confusion protected)
│   │   └── password.go      # HashPassword / CheckPassword (bcrypt)
│   ├── handler/
│   │   └── auth_handler.go  # Login and Me HTTP handlers (depends on the store)
│   ├── middleware/
│   │   └── auth.go          # RequireAuth(): validates the Bearer token, sets userID on context
│   ├── model/
│   │   └── user.go          # User struct (password excluded from JSON)
│   ├── router/
│   │   └── router.go        # Route registration under /api/auth, middleware applied per-route
│   └── store/
│       └── user_store.go    # In-memory user store, seeded with one user
├── .env                     # Local config (not committed)
├── .env.example             # Template for .env
├── go.mod
└── go.sum
```

## Requirements

- Go 1.27+ (see `go.mod`)
- [Postman](https://www.postman.com/) for testing the routes

## Setup

1. Clone the repository and move into the project folder.

2. Create your `.env` from the template:

   ```bash
   cp .env.example .env
   ```

   `.env` holds:

   | Variable              | Description                                   | Example                              |
   | --------------------- | --------------------------------------------- | ------------------------------------ |
   | `JWT_SECRET`          | Secret used to sign and verify tokens (HS256) | `change-me-to-a-long-random-string`  |
   | `JWT_EXPIRES_MINUTES` | Token lifetime in minutes                     | `60`                                 |
   | `PORT`                | Port the server listens on                    | `8080`                               |

   Use a long, random value for `JWT_SECRET`. The `.env` file is git-ignored so the secret is not committed.

3. Install dependencies:

   ```bash
   go mod download
   ```

## Running the server

From the project root:

```bash
go run ./cmd/api
```

The server starts on the port from `.env` (default `8080`). On startup, Gin prints the registered routes and the listening address. Stop the server with `Ctrl+C`.

### Seeded user

The in-memory store is seeded with one user you can use to log in:

| Field    | Value       |
| -------- | ----------- |
| username | `test1`     |
| password | `test123#`  |

## Testing in Postman

> Postman collection / docs: [Postman Collection](https://documenter.getpostman.com/view/57333016/2sBYHNVhmp)

Make sure the server is running (`go run ./cmd/api`) before sending requests. Base URL: `http://localhost:8080`.

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

Copy the `token` value — you'll need it for the next call.

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

### 3. Me — no token

- **Method:** `GET`
- **URL:** `http://localhost:8080/api/auth/me`
- **Authorization:** none.
- **Expected response:** `401 Unauthorized`

  ```json
  {
    "error": "Unauthorized"
  }
  ```

### 4. Me — bogus token

- **Method:** `GET`
- **URL:** `http://localhost:8080/api/auth/me`
- **Authorization:** Bearer Token set to a garbage value like `abc.def.ghi`.
- **Expected response:** `401 Unauthorized`

  ```json
  {
    "error": "Unauthorized"
  }
  ```

### 5. Login — wrong credentials

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
    "error": "invalid credentials"
  }
  ```

  The same `invalid credentials` response is returned whether the username is unknown or the password is wrong, so the API does not reveal which field was incorrect.

## API reference

| Method | Path                 | Auth required | Description                        |
| ------ | -------------------- | ------------- | ---------------------------------- |
| POST   | `/api/auth/login`    | No            | Validate credentials, return a JWT |
| GET    | `/api/auth/me`       | Yes (Bearer)  | Return the current user's profile  |
