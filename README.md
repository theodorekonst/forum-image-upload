# Member 4 — Likes, Filters, Frontend & Delivery

This document tracks my scope of the forum project: the user-facing layer
(HTML templates + CSS), the like/dislike system, the post filters, error
handling, Docker, and testing. It also serves as a running checklist so the
team can see what is done and what is pending.

## Scope

| Area | Responsibility |
|------|----------------|
| Frontend | All HTML templates and CSS styling |
| Likes/Dislikes | Like & dislike buttons for posts and comments |
| Filters | Filter posts by category, by my posts, by liked posts |
| Error pages | 400, 401, 403, 404, 500 |
| Delivery | Dockerfile, optional docker-compose, tests, final integration |

## Dependencies

This work builds on top of the other members' branches:

- **Member 1 (Database & Backend Core):** tables (`users`, `posts`, `comments`,
  `categories`, `post_categories`, `post_likes`, `comment_likes`, `sessions`),
  schema, and the repository/CRUD layer.
- **Member 2 (Auth & Sessions):** registration, login, logout, session cookies,
  password hashing, and `RequireAuth()` middleware that tells the frontend
  whether a user is logged in.
- **Member 3 (Posts, Comments & Categories):** create/view post routes,
  comment creation, category management, and post detail data.

The frontend consumes data these layers produce; the like/filter handlers
read from and write to Member 1's tables and rely on Member 2's session info
to know *who* the current user is.

## Templates

Located in `templates/`. Built with Go's `html/template` (no frontend
frameworks, per the project rules).

| Template | Purpose |
|----------|---------|
| `home.html` | Landing page: lists all posts, shows like/dislike counts and filters |
| `login.html` | Login form (email + password) → `POST /login` |
| `register.html` | Register form (email + username + password) → `POST /register` |
| `post.html` | Single post detail with comments and like/dislike controls |
| `profile.html` | Logged-in user's profile (their posts / liked posts) |
| `error.html` | Generic error page reused for all HTTP error codes |

## Static files

Located in `static/`:

- `static/css/style.css` — site styling (layout, spacing, form/button styles)
- `static/images/` — any icons or images used by the UI

## Likes & Dislikes

Routes handled in `handlers/likes.go`:

- `POST /post/like`
- `POST /comment/like`

Rules:

- Only registered (logged-in) users can like/dislike — enforced via auth
  middleware before the handler runs.
- Like and dislike counts are visible to everyone, including guests.
- A user can have at most one reaction per post/comment.

## Filters

Routes/logic handled in `handlers/filters.go`:

- **By category** — e.g. Programming, Sports, Gaming, Movies (works for everyone).
- **My posts** — posts created by the logged-in user (registered users only).
- **Liked posts** — posts the logged-in user has liked (registered users only).

The last two require a session, since they refer to the currently logged-in
user.

## Error handling

A single `error.html` template is reused for all error responses, with the
correct HTTP status code set by the handler:

- `400` Bad Request
- `401` Unauthorized
- `403` Forbidden
- `404` Not Found
- `500` Internal Server Error

## Docker

- `Dockerfile` — builds and runs the forum container.
- `docker-compose.yml` — optional, for convenience.
- Note: `mattn/go-sqlite3` uses CGO, so the build needs a C toolchain
  available at compile time (this affects how the Dockerfile is written).

## Testing

- `go test ./...`
- Test files live next to the code they test, using the `*_test.go` convention.

## Progress checklist

### Phase 1 — Scaffolding
- [x] Folder structure (`templates/`, `static/css/`, `static/images/`)
- [x] Merge Member 1 & Member 2 branches into this branch
- [x] `templates/login.html` (basic structure)
- [x] `templates/register.html` (basic structure)
- [ ] Placeholder `handlers/likes.go`
- [ ] Placeholder `handlers/filters.go`

### Phase 2 — Core pages & styling
- [ ] `templates/home.html`
- [ ] `static/css/style.css` (base styling)
- [ ] Wire login/register forms to their routes (verify end-to-end)

### Phase 3 — Likes & errors
- [ ] Like/dislike for posts
- [ ] Like/dislike for comments
- [ ] `templates/error.html`
- [ ] Error responses for 400/401/403/404/500

### Phase 4 — Filters & delivery
- [ ] Category filter
- [ ] My posts filter
- [ ] Liked posts filter
- [ ] `Dockerfile`
- [ ] Tests for likes & filters
- [ ] Final front-end polish

### Phase 5 — Integration
- [ ] Full team integration & delivery checklist verified