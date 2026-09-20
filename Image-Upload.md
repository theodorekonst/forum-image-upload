# Image Upload Optional - Project Plan

This document describes the structure, branches, tasks, and file changes required to extend
the forum so registered users can attach an image to a post.

The goal is to support:

- Uploading a PNG, JPEG, or GIF image when creating a post
- Rejecting images over 20MB with a clear error message
- Rejecting file types other than PNG, JPEG, and GIF
- Showing the image to everyone (registered users and guests) when viewing the post
- Keeping the image after navigating away and back

---

# Branch Structure

```text
main            → final submission + documentation
dev             → integration + image upload implementation + testing
```

---

# Task Structure (4 Tasks)

## Task 1 - Database

### Responsibilities

- Add columns to store the image with its post
- Update the `Post` model to match

### Files to Create / Modify

```text
db/schema.sql
db/repo.go
```

## Task 2 - Reading and Validating the Upload

### Responsibilities

- Change the create-post form to send files (`multipart/form-data`)
- Read the uploaded file with `r.FormFile("image")`
- Enforce the 20MB limit with `http.MaxBytesReader`
- Detect the real file type with `http.DetectContentType`
- Accept only `image/png`, `image/jpeg`, `image/gif`
- Show a clear error for oversized or unsupported files
- Treat the image as optional (a post can still be created without one)

### Files to Create / Modify

```text
templates/create_post.html
handlers/posts.go
```

## Task 3 - Saving and Displaying the Image

### Responsibilities

- Save the image bytes, type, and size together with the post
- Add a route that serves the image bytes for a given post
- Show the image on the post page (and home feed, if applicable)

### Files to Create / Modify

```text
db/repo_extra.go
cmd/server/main.go
handlers/posts.go
templates/post.html
templates/home.html
```

## Task 4 - Integration & Testing (dev)

### Responsibilities

- Verify PNG, JPEG, and GIF uploads are accepted
- Verify an image over 20MB is rejected with a message, not a crash
- Verify a renamed non-image file is rejected
- Verify the image is still visible after navigating away and back
- Verify a post can still be created without an image
- Verify deleting a post does not leave the site in a broken state

### Files to Modify

```text
db/repo_test.go
handlers/content_validation_test.go
handlers/posts_test.go   (NEW)
README.md
```

# Suggested Project Structure

```text
forum-image-upload/
├── cmd/
│   └── server/main.go
├── db/
│   ├── schema.sql
│   ├── repo.go
│   ├── repo_extra.go
│   ├── repo_test.go
│   └── auth_test.go
├── handlers/
│   ├── posts.go
│   ├── posts_test.go        (NEW)
│   ├── content_validation_test.go
│   └── errors.go
├── templates/
│   ├── create_post.html
│   ├── post.html
│   └── home.html
├── .gitignore
├── build.sh
├── README.md
└── Image-Upload.md
```
