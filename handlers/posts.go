package handlers

import (
	"errors"
	"forum/db"
	"forum/middleware"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// maxImageSize is the largest image accepted on a post, per the image-upload subject.
const maxImageSize = 20 << 20 // 20 MB

func CreatePostHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.GetUserFromRequest(store, r)
		if r.Method == http.MethodGet {
			categories, err := store.ListCategories()
			if err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not load categories.")
				return
			}
			if err := renderPage(w, "create_post.html", map[string]any{
				"User":       user,
				"Categories": categories,
			}); err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not render the create post page.")
			}
			return

		}
		if r.Method == http.MethodPost {
			r.Body = http.MaxBytesReader(w, r.Body, maxImageSize)
			isMultipart := true
			switch err := r.ParseMultipartForm(maxImageSize); {
			case err == nil:
				// multipart form parsed, image (if any) is available via r.FormFile.
			case errors.Is(err, http.ErrNotMultipart):
				// plain application/x-www-form-urlencoded request, no image possible.
				isMultipart = false
				if err := r.ParseForm(); err != nil {
					RenderError(w, http.StatusBadRequest, "Could not read the form.")
					return
				}
			default:
				categories, cErr := store.ListCategories()
				if cErr != nil {
					RenderError(w, http.StatusInternalServerError, "Could not load categories.")
					return
				}
				w.WriteHeader(http.StatusBadRequest)
				renderPage(w, "create_post.html", map[string]any{
					"User":       user,
					"Categories": categories,
					"ImageError": "Image is too big. Max size is 20MB.",
					"FirstError": "Image",
				})
				return
			}

			title := strings.TrimSpace(r.FormValue("title"))
			content := strings.TrimSpace(r.FormValue("content"))
			categoryValues := r.Form["categories"]

			var imageData []byte
			var imageType string
			var imageSize int64
			if isMultipart {
				if file, header, ferr := r.FormFile("image"); ferr == nil {
					defer file.Close()
					buf, rerr := io.ReadAll(file)
					if rerr != nil {
						RenderError(w, http.StatusInternalServerError, "Could not read the uploaded image.")
						return
					}
					detectedType := http.DetectContentType(buf)
					if detectedType != "image/png" && detectedType != "image/jpeg" && detectedType != "image/gif" {
						categories, cErr := store.ListCategories()
						if cErr != nil {
							RenderError(w, http.StatusInternalServerError, "Could not load categories.")
							return
						}
						w.WriteHeader(http.StatusBadRequest)
						renderPage(w, "create_post.html", map[string]any{
							"User":       user,
							"Categories": categories,
							"Title":      title,
							"Content":    content,
							"ImageError": "Only PNG, JPEG, and GIF images are allowed.",
							"FirstError": "Image",
						})
						return
					}
					imageData = buf
					imageType = detectedType
					imageSize = header.Size
				} else if !errors.Is(ferr, http.ErrMissingFile) {
					RenderError(w, http.StatusBadRequest, "Could not read the uploaded image.")
					return
				}
			}

			if title == "" || content == "" || len(categoryValues) == 0 {
				categories, err := store.ListCategories()
				if err != nil {
					RenderError(w, http.StatusInternalServerError, "Could not load categories.")
					return
				}
				data := map[string]any{
					"User":       user,
					"Categories": categories,
					"Title":      title,
					"Content":    content,
				}
				if title == "" {
					data["TitleError"] = "Title cannot be empty."
					data["FirstError"] = "Title"
				}
				if content == "" {
					data["ContentError"] = "Content cannot be empty."
					if _, exists := data["FirstError"]; !exists {
						data["FirstError"] = "Content"
					}
				}
				if len(categoryValues) == 0 {
					data["CategoryError"] = "Choose at least one category."
					if _, exists := data["FirstError"]; !exists {
						data["FirstError"] = "Category"
					}
				}
				w.WriteHeader(http.StatusBadRequest)
				if err := renderPage(w, "create_post.html", data); err != nil {
					return
				}
				return
			}
			categoryIDs := make([]int64, 0, len(categoryValues))
			for _, cidStr := range categoryValues {
				cid, err := strconv.ParseInt(cidStr, 10, 64)
				if err != nil || cid <= 0 {
					RenderError(w, http.StatusBadRequest, "Invalid category.")
					return
				}
				categoryIDs = append(categoryIDs, cid)
			}
			postID, err := store.CreatePostWithCategories(user.ID, title, content, categoryIDs, imageData, imageType, imageSize)
			if errors.Is(err, db.ErrCategoryRequired) || errors.Is(err, db.ErrInvalidCategory) {
				RenderError(w, http.StatusBadRequest, "Invalid category.")
				return
			}
			if err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not create the post.")
				return
			}
			http.Redirect(w, r, "/post?id="+strconv.FormatInt(postID, 10), http.StatusSeeOther)
			return
		}

		RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
	}
}

// PostImageHandler serves the image bytes stored with a post.
func PostImageHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
			return
		}

		id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		if err != nil {
			RenderError(w, http.StatusBadRequest, "Invalid post ID.")
			return
		}

		post, err := store.GetPostByID(id)
		if errors.Is(err, db.ErrNotFound) {
			RenderError(w, http.StatusNotFound, "Post not found.")
			return
		}
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not load the post.")
			return
		}
		if len(post.ImageData) == 0 {
			RenderError(w, http.StatusNotFound, "This post has no image.")
			return
		}

		w.Header().Set("Content-Type", post.ImageType)
		w.Write(post.ImageData)
	}
}

type viewComment struct {
	ID           int64
	AuthorID     int64
	Author       string
	Content      string
	CreatedAt    time.Time
	IsDeleted    bool
	Likes        int64
	Dislikes     int64
	UserReaction string
	Replies      []viewComment
}

type postPageData struct {
	User             *db.User
	Post             *db.Post
	PostAuthor       string
	PostCategories   []string
	PostLikes        int64
	PostDislikes     int64
	PostUserReaction string
	Comments         []viewComment
	ScrollTo         string
	CommentError     string
}

func ViewPostHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
			return
		}

		user, _ := middleware.GetUserFromRequest(store, r)

		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			RenderError(w, http.StatusBadRequest, "Invalid post ID.")
			return
		}

		post, err := store.GetPostByID(id)
		if errors.Is(err, db.ErrNotFound) {
			RenderError(w, http.StatusNotFound, "Post not found.")
			return
		}
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not load the post.")
			return
		}

		postAuthor, err := store.GetUserByID(post.AuthorID)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not load post author.")
			return
		}

		postCategories, err := store.ListCategoriesByPost(post.ID)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not load post categories.")
			return
		}

		postLikes, err := store.CountPostLikes(post.ID)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not load reactions.")
			return
		}
		postDislikes, err := store.CountPostDislikes(post.ID)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not load reactions.")
			return
		}
		postUserReaction := ""
		if user != nil {
			postUserReaction, err = store.GetPostReaction(post.ID, user.ID)
			if err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not load reactions.")
				return
			}
		}

		rawComments, err := store.ListCommentsByPost(id)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not load comments.")
			return
		}

		// Build a flat map of all view comments first
		vcMap := make(map[int64]viewComment, len(rawComments))
		for _, c := range rawComments {
			authorName := "[deleted]"
			content := "[deleted]"
			var likes, dislikes int64
			userReaction := ""
			if !c.IsDeleted {
				commentAuthor, err := store.GetUserByID(c.AuthorID)
				if err != nil {
					RenderError(w, http.StatusInternalServerError, "Could not load comment author.")
					return
				}
				authorName = commentAuthor.Username
				content = c.Content
				likes, err = store.CountCommentLikes(c.ID)
				if err != nil {
					RenderError(w, http.StatusInternalServerError, "Could not load reactions.")
					return
				}
				dislikes, err = store.CountCommentDislikes(c.ID)
				if err != nil {
					RenderError(w, http.StatusInternalServerError, "Could not load reactions.")
					return
				}
				if user != nil {
					userReaction, err = store.GetCommentReaction(c.ID, user.ID)
					if err != nil {
						RenderError(w, http.StatusInternalServerError, "Could not load reactions.")
						return
					}
				}
			}
			vcMap[c.ID] = viewComment{
				ID:           c.ID,
				AuthorID:     c.AuthorID,
				Author:       authorName,
				Content:      content,
				CreatedAt:    c.CreatedAt,
				IsDeleted:    c.IsDeleted,
				Likes:        likes,
				Dislikes:     dislikes,
				UserReaction: userReaction,
			}
		}

		// Attach replies to their parent, collect top-level comments in order
		var comments []viewComment
		for _, c := range rawComments {
			if c.ParentID != nil {
				continue // handled as reply
			}
			vc := vcMap[c.ID]
			for _, r := range rawComments {
				if r.ParentID != nil && *r.ParentID == c.ID {
					vc.Replies = append(vc.Replies, vcMap[r.ID])
				}
			}
			comments = append(comments, vc)
		}

		// Accept only known targets to prevent reflected XSS via the scroll param.
		scrollTo := ""
		if raw := r.URL.Query().Get("scroll"); raw == "comment-form" {
			scrollTo = raw
		} else if len(raw) > 8 && raw[:8] == "comment-" {
			if _, err := strconv.ParseInt(raw[8:], 10, 64); err == nil {
				scrollTo = raw
			}
		}
		commentError := ""
		if r.URL.Query().Get("comment_error") == "required" {
			commentError = "Comment text cannot be empty."
		}

		if err := renderPage(w, "post.html", postPageData{
			User:             user,
			Post:             post,
			PostAuthor:       postAuthor.Username,
			PostCategories:   postCategories,
			PostLikes:        postLikes,
			PostDislikes:     postDislikes,
			PostUserReaction: postUserReaction,
			Comments:         comments,
			ScrollTo:         scrollTo,
			CommentError:     commentError,
		}); err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not render the post page.")
		}
	}
}
