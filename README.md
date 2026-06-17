# Forum Authentication & Session Management (Member 2)

This README covers Member 2's responsibilities implemented so far: registration, login, logout placeholder, password hashing, validation, error handling, session lookup, and authentication middleware.

## Member 2 scope (current)

- Registration handler: handlers/auth.go
- Login handler: handlers/auth.go
- Logout placeholder: handlers/auth.go
- Server setup & routes: cmd/server/main.go
- Session lookup helper: middleware/auth.go (GetUserFromRequest)
- Auth middleware: middleware/auth.go (RequireAuth / RequireGuest)
- Includes Member 1's branch repository and uses: db/repo.go

## Quick setup & commands

- Install dependencies:

```bash
go mod tidy
```

- Run the server:

```bash
go run ./cmd/server
```

- Test endpoints (example):

```bash
curl -d "email=test@a.com&username=user123&password=Password1!" http://localhost:8080/register
curl -d "email=test@a.com&password=Password1!" http://localhost:8080/login
sqlite3 forum.db "SELECT * FROM users;"
```

## Available routes

- GET/POST /register
- GET/POST /login
- GET/POST /logout (placeholder)

## What works now:

- Input validation (email, username, password)
- bcrypt password hashing
- Login credential verification
- Unified handler structure
- Initial route setup
- Session lookup from cookie (GetUserFromRequest)
- Auth middleware (RequireAuth / RequireGuest)

## TODO / Next steps:

- Integrate session creation into login
- Integrate session deletion into logout
- Secure cookie handling
- Access control for protected routes (after session integration)
- Unit test files for all auth functions
- Final README update after full integration
