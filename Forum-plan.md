# Recommended Team Structure

## Member 1 — Database & Backend Architecture (Foundation)

### Responsibilities

* SQLite database design
* ERD (Entity Relationship Diagram)
* Database initialization
* SQL queries
* Repository layer

### Deliverables

* `forum.db`
* Database schema
* Migration/creation scripts
* CRUD functions

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

## Member 2 — Authentication & Sessions

### Responsibilities

* Registration
* Login
* Logout
* Session management
* Cookies
* Password hashing (bcrypt)

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

* User identity for Members 3, 4 and 5

---

## Member 3 — Posts, Comments & Categories

### Responsibilities

* Create posts
* View posts
* Create comments
* Categories
* Post details page

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

## Member 4 — Likes, Dislikes & Filters

### Responsibilities

* Like post
* Dislike post
* Like comment
* Dislike comment
* Filter system

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

### Priority

**Medium-High**

Depends on:

* Posts
* Authentication

---

## Member 5 — Frontend, Error Handling, Docker & Testing

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

# Development Flow

## Phase 1 (Days 1–2)

### Member 1

Database schema

### Member 5

HTML structure
Folder architecture

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

### Member 2

Authentication

### Member 3

Posts & comments

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

### Member 4

Likes/dislikes

### Member 5

Frontend styling

Result:

```text
Like Post
Dislike Post
Like Comment
Dislike Comment
```

---

## Phase 4 (Days 5–6)

### Member 4

Filters

Result:

```text
Category Filter
My Posts Filter
Liked Posts Filter
```

---

## Phase 5 (Days 6–7)

### Entire Team

Integration

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
