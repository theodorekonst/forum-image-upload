# Recommended Team Structure (4 Members)

## Member 1 — Database & Backend Core

### Responsibilities

* SQLite database design
* ERD (Entity Relationship Diagram)
* Database initialization and migrations
* SQL queries and repository layer
* Core CRUD functions for users, posts, comments and categories

### Deliverables

* `forum.db`
* Database schema
* Migration/creation scripts
* CRUD helpers for all core entities

### Tables

```sql
users
posts
comments
categories
post_categories
post_likes
comment_likes
sessions
```

### Priority

**Critical**

Everything depends on this work.

---

## Member 2 — Authentication & Session Management

### Responsibilities

* Registration
* Login
* Logout
* Session handling and cookies
* Password hashing (bcrypt)
* Auth middleware and access control

### Deliverables

#### Registration

```text
email
username
password
```

#### Login

```text
email
password
```

#### Session Cookie

```go
http.SetCookie(...)
```

#### Middleware

```go
RequireAuth()
```

### Priority

**Critical**

Depends on:

* Database tables from Member 1

Provides:

* User identity for Members 3 and 4

---

## Member 3 — Posts, Comments & Categories

### Responsibilities

* Create and view posts
* Create comments
* Manage categories
* Build post detail pages
* Support the main content flow for the forum

### Deliverables

#### Create Post

```text
Title
Content
Categories
```

#### Create Comment

```text
Comment text
```

#### Routes

```go
POST /post/create
GET /post?id=1
POST /comment/create
```

### Priority

**Critical**

Depends on:

* Database
* Authentication

---

## Member 4 — Likes, Filters, Frontend & Delivery

### Responsibilities

* Like/dislike posts and comments
* Filter posts by category, user, and liked items
* Build HTML templates and CSS styling
* Handle error pages and final integration
* Docker setup, testing and release readiness

### Deliverables

### Likes

```go
POST /post/like
POST /comment/like
```

### Filters

#### By Category

```text
Programming
Sports
Gaming
Movies
```

#### By User Posts

```text
My Posts
```

#### By Liked Posts

```text
Liked Posts
```

### Pages

```text
Home
Login
Register
Create Post
Post Details
Profile
Error Pages
```

### Quality & Delivery

```text
400
401
403
404
500
Dockerfile
Docker Compose (optional)
go test ./...
```

### Priority

**High**

Depends on:

* Posts
* Authentication
* Final integration from the full team

---

### Responsibilities

* HTML templates
* CSS
* Error pages
* Docker
* Testing
* Final integration

### Deliverables

### Pages

```text
Home
Login
Register
Create Post
Post Details
Profile
Error Pages
```

### Errors

```text
400
401
403
404
500
```

### Docker

```dockerfile
Dockerfile
docker-compose.yml (optional)
```

### Testing

```go
go test ./...
```

### Priority

**Critical for final delivery**

Works in parallel with everyone.

---

# Development Flow (4 Members)

## Phase 1 (Days 1–2)

### Member 1

Database schema and repository foundation

### Member 2

Authentication skeleton and session setup

### Member 3

Core routes and models for posts, comments and categories

### Member 4

Project scaffolding, templates, static files and initial UI structure

Result:

```text
project/
├── cmd/
├── database/
├── handlers/
├── models/
├── templates/
├── static/
├── Dockerfile
└── main.go
```

---

## Phase 2 (Days 2–4)

### Member 1

Complete CRUD and database integration

### Member 2

Registration, login, logout and auth middleware

### Member 3

Post creation, post view and comment creation

### Member 4

Home page, auth pages and basic styling

Result:

```text
Register
Login
Logout
Create Post
View Post
Create Comment
```

---

## Phase 3 (Days 4–5)

### Member 1

Support fixes and data consistency improvements

### Member 2

Session validation and user access checks

### Member 3

Category handling and post detail experience

### Member 4

Likes, dislikes, error pages and quality checks

Result:

```text
Like Post
Dislike Post
Like Comment
Dislike Comment
```

---

## Phase 4 (Days 5–6)

### Member 1

Final database verification and query optimization

### Member 2

Security review and auth hardening

### Member 3

Post/comment polish and category flow refinement

### Member 4

Filters, Docker setup, testing and final front-end polish

Result:

```text
Category Filter
My Posts Filter
Liked Posts Filter
```

---

## Phase 5 (Days 6–7)

### Entire Team

Final integration and delivery

Checklist:

* Register works
* Login works
* Sessions work
* Posts work
* Comments work
* Categories work
* Likes work
* Dislikes work
* Filters work
* Error pages work
* Docker works

---

# Suggested Project Structure

```text
forum/
│
├── cmd/
│   └── main.go
│
├── database/
│   ├── schema.sql
│   └── sqlite.go
│
├── handlers/
│   ├── auth.go
│   ├── posts.go
│   ├── comments.go
│   ├── likes.go
│   └── filters.go
│
├── middleware/
│   └── auth.go
│
├── models/
│   ├── user.go
│   ├── post.go
│   ├── comment.go
│   └── category.go
│
├── templates/
│   ├── home.html
│   ├── login.html
│   ├── register.html
│   ├── post.html
│   └── error.html
│
├── static/
│   ├── css/
│   └── images/
│
├── tests/
│
├── Dockerfile
├── go.mod
└── forum.db
```

# Final Priority Order

1. Database Schema
2. Authentication & Sessions
3. Posts
4. Comments
5. Categories
6. Likes/Dislikes
7. Filters
8. Error Handling
9. Docker
10. Testing

This order minimizes blocking between team members and ensures that everyone can work in parallel while integrating smoothly at the end.
