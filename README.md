# Member 4 - Frontend, Reactions, Filters, and Delivery

## Responsibilities

| Area | Member 4 responsibility |
|---|---|
| Frontend | HTML templates, shared CSS, logo assets, and responsive layout |
| Home page | Public feed, category navigation, user-specific filters, and UI states |
| Reactions | Like/dislike behavior for posts and comments |
| Comments | Nested/threaded comment system with inline reply forms |
| Errors | Shared error page and handler integration for 400, 401, 403, 404, and 500 |
| Delivery | Dockerfile, build script, tests, final integration, and release verification |

## Work Completed

- Built shared forum styling, logo integration, and post reaction controls.
- Added public home-page rendering with category, my-posts, and liked-posts filters.
- Added template rendering helpers and template parsing tests.
- Integrated shared styling into home, login, register, post-detail, and create-post pages.
- Added shared error page helper and tests.
- Integrated static asset serving and corrected normal page rendering for login and registration.
- Merged post/comment work and resolved template-loading integration.
- Implemented full like/dislike reaction system for posts and comments (schema, endpoints, toggle logic, UI).
- Added inline form error messages for register and login.
- Added 404 handling for unknown routes.
- Added Dockerfile with multi-stage build and `build.sh` script for containerized deployment.
- Fixed post detail page navigation to be auth-aware (login/logout based on session).
- Implemented nested/threaded comments with `parent_id` FK, tree-building handler, and inline reply forms.
- Added AJAX reactions via `reactions.js` — optimistic UI updates with no page reload.
- Added relative timestamps via `timeago.js` (`just now`, `3 minutes ago`, etc.).
- Fixed liked-posts filter to correctly show only liked posts (not disliked).
- Fixed scroll position after posting a comment or reply — page anchors to the new comment.
- Implemented sticky navbar so it remains visible while scrolling.
- Comprehensive dark SaaS design system: Hanken Grotesk font, card hierarchy, feed container, category dot, sidebar active states, discussion thread UI, and post detail layout.

## Member 4 Files

- `templates/` — home, login, register, post detail, create post, and error UI
- `static/css/style.css` — shared site styling and design system
- `static/images/` — forum logo assets
- `static/js/reactions.js` — AJAX like/dislike with optimistic UI
- `static/js/timeago.js` — relative timestamp rendering
- `handlers/home.go` — public home feed, filters, and excerpt logic
- `handlers/errors.go` — shared error renderer
- `handlers/templates.go` — common template rendering helper
- `handlers/reactions.go` — like/dislike handlers for posts and comments
- `handlers/comments.go` — comment and reply creation with scroll anchor redirect
- `handlers/posts.go` — post detail with nested comment tree and scroll target
- `db/schema.sql` — comments table with `parent_id` for threading
- `db/repo_extra.go` — reaction queries, liked-posts filter, comment with parent
- `Dockerfile` — multi-stage Docker build
- `build.sh` — Docker build and run script
