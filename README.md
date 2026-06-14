# Forum Backend Core (Member 1)

This repository contains the backend database core for the forum application. This README focuses on Member 1 responsibilities: schema, migrations, repository layer, seeding, and tests.

**Member 1 scope**

- **Schema & Migrations**: [db/schema.sql](db/schema.sql)
- **DB Initialization & Migrations**: [db/db.go](db/db.go)
- **Repository (CRUD)**: [db/repo.go](db/repo.go) and [db/repo_extra.go](db/repo_extra.go)
- **Seeding CLI**: [cmd/seed/main.go](cmd/seed/main.go)
- **Unit tests**: [db/repo_test.go](db/repo_test.go)

**Quick setup & commands**

- Ensure modules and dependencies:

```bash
go mod tidy
```

- Seed the database (creates `forum.db` in repo root):

```bash
go run ./cmd/seed
```

- Run repository unit tests:

```bash
go test ./...
```

**Implementation notes**

- Uses `github.com/mattn/go-sqlite3` with the SQLite driver name `sqlite3`.
- `go-sqlite3` requires CGO to be enabled when building or running tests.
- The SQL schema is embedded in the `db` package so migrations run reliably from tests and CLIs.
- `db/repo.go` implements core CRUD for `users`, `posts`, `comments`, `categories`, and `sessions`.
- `db/repo_extra.go` provides likes and filtered query helper methods.

**Notes**

- If `go test ./...` fails due to CGO, install a C compiler and run with `CGO_ENABLED=1`.
- The repository is intentionally scoped to Member 1 database backend work only.

**Next recommended Member 1 tasks**

- Add migration/versioning tooling (optional).
- Expand unit tests to cover failure cases and constraint enforcement.
- Add a small CLI for DB inspection/export if needed.

