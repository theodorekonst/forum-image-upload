# Authentication Optional - Project Plan

This document describes the structure, branches, tasks, and file changes required to extend
the forum with GitHub and Google OAuth authentication.

The goal is to support:

- Login with GitHub
- Login with Google
- Unified user model (password + OAuth)
- Full session integration
- Audit compliance

---

# Branch Structure

```text
main            → final submission + documentation
dev             → integration + testing
oauth-github    → GitHub OAuth implementation
oauth-google    → Google OAuth implementation
```

Each OAuth provider is implemented in its own branch, merged into `dev` for testing,
and finally merged into `main` when stable.

---

# Task Structure (3 Tasks)

## Task 1 - GitHub OAuth

### Responsibilities

- Redirect to GitHub OAuth
- Handle callback
- Exchange code → access token
- Fetch GitHub user info
- Create or fetch OAuth user
- Create session + cookie
- Redirect to /

### Files to Create / Modify

```text
handlers/oauth_github.go      (NEW)
handlers/oauth_handlers.go
cmd/server/main.go
templates/login.html
templates/register.html
db/repo.go
db/schema.sql
```

## Task 2 - Google OAuth

### Responsibilities

- Redirect to Google OAuth
- Handle callback
- Exchange code → access token
- Fetch Google user info
- Create or fetch OAuth user
- Create session + cookie
- Redirect to /

### Files to Create / Modify

```text
handlers/oauth_google.go      (NEW)
handlers/oauth_handlers.go
cmd/server/main.go
templates/login.html
templates/register.html
db/repo.go
db/schema.sql
```

## Task 3 - Integration & Testing (dev)

### Responsibilities

- Merge GitHub + Google OAuth branches
- Ensure OAuth users can:
  - create posts
  - comment
  - react
  - filter posts
  - log out / log in
- Ensure duplicate OAuth registration fails
- Ensure normal registration still works
- Ensure sessions work for OAuth users

### Files to Modify

```text
handlers/oauth_handlers.go
middleware/auth.go
handlers/
auth_test.go
handlers/oauth_test.go     (NEW)
```

# Suggested Project Structure

```text
forum/
├── cmd/
│   └── server/main.go
├── db/
│   ├── schema.sql
│   ├── repo.go
│   └── auth_test.go
├── handlers/
│   ├── auth.go
│   ├── auth_github.go       (NEW)
│   ├── auth_google.go       (NEW)
│   ├── auth_test.go
│   ├── oauth_test.go        (NEW)
│   └── oauth_handlers.go    (NEW)
├── middleware/
│   └── auth.go
├── handlers/
│   ├── css/styles.css
│   └── images/
├── templates/
│   ├── register.html
│   └── login.html
├── .env.example
├── .gitignore
├── build.sh
├── README.md
└── Authentication.md
```
