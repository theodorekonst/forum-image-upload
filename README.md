# Forum

A full-stack web forum built with Go, SQLite, HTML, CSS, and vanilla
JavaScript. Visitors can browse posts, comments, categories, and reaction
counts. Registered users can publish content, reply to discussions, react to
posts and comments, and filter the feed by their own activity.

The project uses only Go's standard library and the packages allowed by the
forum subject. No frontend framework is used.

## Features

### Authentication and sessions

- Registration with email, username, password, and password confirmation
- Unique email and username validation
- Password hashing with `bcrypt`
- Generic login errors that do not reveal whether an account exists
- UUID-based sessions stored in SQLite
- `HttpOnly`, `SameSite=Lax` cookies with a 24-hour expiration
- One active session per user: a new login invalidates previous sessions
- Guest-only login and registration pages
- OAuth login support for GitHub, Google and Discord
- State validation and secure callback handling for OAuth providers

### Posts, comments, and categories

- Public post feed and post detail pages
- Authenticated post and comment creation
- Server-side validation for empty or whitespace-only content
- One or more categories required for every post
- Atomic post and category creation using an SQLite transaction
- Nested comment replies
- Author-only deletion for posts and comments
- Comment tombstones when a deleted comment has replies
- Cascade cleanup of related comments, reactions, and category links

The default categories are:

- Programming
- Sports
- Gaming
- Movies
- General

They are inserted automatically when a new database is initialized.

### Reactions and filters

- Like or dislike posts and comments
- One reaction per user and resource
- Selecting the current reaction again removes it
- Switching reaction replaces the previous one
- Public like, dislike, and comment counts
- AJAX reaction updates with a server-rendered fallback
- Feed filters for:
  - category
  - posts created by the logged-in user
  - posts liked by the logged-in user

### User experience and error handling

- Responsive interface built with HTML, CSS, and vanilla JavaScript
- Accessible inline form validation
- Relative timestamps
- Confirmation dialog for destructive actions
- Shared error pages for `400`, `401`, `403`, `404`, `405`, and `500`
- Correct HTTP method checks on application routes

## Technology

- Go 1.25
- SQLite through `github.com/mattn/go-sqlite3`
- bcrypt through `golang.org/x/crypto/bcrypt`
- UUID through `github.com/google/uuid`
- Go `net/http` and `html/template`
- HTML, CSS, and vanilla JavaScript
- Docker with a multi-stage Alpine build

## Run with Docker

Docker is the recommended way to run the project.

### Build and start with the provided script

```bash
./build.sh
```

The script:

1. builds an image named `forum`;
2. replaces an existing container named `forum`;
3. starts the application on port `8080`.

Open [http://localhost:8080](http://localhost:8080).

### Run the commands manually

```bash
docker build -t forum .
docker run --name forum -p 8080:8080 --env-file .env forum
```

The container expects the required OAuth variables from `.env`, including
`APP_BASE_URL`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`,
`GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `DISCORD_CLIENT_ID`, and
`DISCORD_CLIENT_SECRET`.

The container creates `forum.db` and initializes the schema and default
categories on first startup.

Useful commands:

```bash
docker ps
docker logs forum
docker stop forum
docker start forum
```

> Removing the container also removes the database stored inside that
> container. Keep the container if you want to retain its local forum data.

## Run locally

### Requirements

- Go 1.25 or newer
- A C compiler required by `go-sqlite3`/CGO
- SQLite CLI, optional, for inspecting the database

From the project root:

```bash
go mod download
go run ./cmd/server
```

Then open [http://localhost:8080](http://localhost:8080).

The application creates `forum.db` automatically in the current working
directory. Database files are ignored by Git.

To inspect the database:

```bash
sqlite3 forum.db
```

Example queries:

```sql
SELECT id, email, username, created_at FROM users;
SELECT id, author_id, title, created_at FROM posts;
SELECT id, post_id, author_id, content, created_at FROM comments;
```

## Tests and verification

Run the complete test suite:

```bash
go test ./...
```

Run static analysis:

```bash
go vet ./...
```

Build the server without running it:

```bash
mkdir -p bin
go build -o bin/forum ./cmd/server
```

The tests cover repository operations, authentication, sessions, middleware,
template parsing, validation, ownership rules, reactions, filtering-related
queries, transactional post creation, deletion cascades, and comment
tombstones.

## Application routes

| Method | Route | Access | Purpose |
|---|---|---|---|
| `GET` | `/` | Public | View and filter posts |
| `GET`, `POST` | `/register` | Guests | Register an account |
| `GET`, `POST` | `/login` | Guests | Log in |
| `POST` | `/logout` | Registered | End the active session |
| `GET` | `/auth/github` | Public | Start GitHub OAuth login |
| `GET` | `/auth/github/callback` | Public | Handle GitHub OAuth callback |
| `GET` | `/auth/google` | Public | Start Google OAuth login |
| `GET` | `/auth/google/callback` | Public | Handle Google OAuth callback |
| `GET` | `/auth/discord` | Public | Start Discord OAuth login |
| `GET` | `/auth/discord/callback` | Public | Handle Discord OAuth callback |
| `GET` | `/post?id={id}` | Public | View a post and its comments |
| `GET`, `POST` | `/post/create` | Registered | Display or submit the post form |
| `POST` | `/post/delete` | Post author | Delete a post |
| `POST` | `/post/react` | Registered | Like, dislike, or clear a post reaction |
| `POST` | `/comment/create` | Registered | Create a comment or reply |
| `POST` | `/comment/delete` | Comment author | Delete a comment |
| `POST` | `/comment/react` | Registered | Like, dislike, or clear a comment reaction |
| `GET` | `/static/*` | Public | Serve CSS, JavaScript, and images |

User-specific filters redirect guests to the login page:

```text
/?filter=mine
/?filter=liked
```

Category filters are public:

```text
/?category=Programming
```

## Database design

The SQLite schema is defined in `db/schema.sql` and initialized automatically.
Foreign-key enforcement is enabled for every database connection.

Main tables:

| Table | Purpose |
|---|---|
| `users` | Accounts and bcrypt password hashes |
| `sessions` | UUID session tokens and expiration timestamps |
| `posts` | Forum posts and their authors |
| `comments` | Comments, replies, and deletion tombstones |
| `categories` | Available post categories |
| `post_categories` | Many-to-many relationship between posts and categories |
| `post_likes` | One like or dislike per user and post |
| `comment_likes` | One like or dislike per user and comment |

Important constraints include unique emails, usernames, and session tokens;
foreign keys with cascade cleanup; reaction type checks; and unique
user/resource reaction pairs.

## Project structure

```text
.
├── cmd/
│   ├── server/       # Web server entry point
│   └── seed/         # Optional local development data command
├── db/               # SQLite initialization, schema, repositories, and tests
├── handlers/         # HTTP handlers and handler tests
├── middleware/       # Authentication middleware and tests
├── static/
│   ├── css/          # Shared styling
│   ├── images/       # Logo assets
│   └── js/           # Validation, reactions, timestamps, and dialogs
├── templates/        # Go HTML templates
├── Dockerfile        # Multi-stage container build
├── build.sh          # Docker build and run helper
├── go.mod
└── go.sum
```

## Security notes

- Passwords are never stored in plain text.
- SQL queries use placeholders rather than string concatenation.
- Go templates escape user-generated HTML.
- Session cookies are inaccessible to JavaScript.
- Login failures use a generic response.
- Protected actions require authentication and use non-`GET` methods.
- Post and comment deletion verifies resource ownership.
- Database foreign keys preserve relational integrity.

For an HTTPS production deployment, set the session cookie's `Secure`
attribute to `true`.

## Contributors

This project was developed collaboratively. Individual contributions can be
reviewed in the Git history.