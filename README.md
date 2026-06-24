# Forum Authentication & Session Management (Member 2)

This README covers Member 2's responsibilities: registration, login, logout, session handling, cookies, bcrypt hashing, auth middleware, and access control.

## Member 2 scope (current)

- Registration (POST /register)
- Login (POST /login)
- Logout (POST /logout)
- Session creation & deletion
- Single active session per user
- Cookie‑based authentication
- Auth middleware (RequireAuth / RequireGuest)
- Mux router integration
- Uses Member 1's repository layer (db/repo.go)

## Quick setup & commands

- Install dependencies:

```bash
go mod tidy
```

- Run the server:

```bash
go run ./cmd/server
```

- Test examples:

  **Register:** `curl -d "email=test1@example.com&username=user123&password=Password_123" http://localhost:8080/register`
  **Login:** `curl -d "email=test1@example.com&password=Password_123" http://localhost:8080/login`
  **Check DB:** `sqlite3 forum.db "SELECT * FROM sessions;"`
  **Expire session:** `UPDATE sessions SET expires_at='2000-01-01 00:00:00' WHERE id=X;`
  **Test cookie:** `curl -i -b "session=TOKEN" http://localhost:8080/`
  **Logout:** `curl -i -b "session=TOKEN" http://localhost:8080/logout`

## Available routes

- GET/POST /register
- GET/POST /login
- POST /logout
- GET / (home page, protected)

## What works now:

- Input validation (email, username, password)
- bcrypt password hashing
- Session storage with local timezone and expiration
- One active session per user
- Cookie handling (HttpOnly, SameSite)
- Protected routes via middleware
- Clean server routing with mux
