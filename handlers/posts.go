package handlers

import (
	"forum/db"
	"forum/middleware"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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
			title := strings.TrimSpace(r.FormValue("title"))
			content := strings.TrimSpace(r.FormValue("content"))
			categoryIDs := r.Form["categories"]
			if title == "" || content == "" {
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
				w.WriteHeader(http.StatusBadRequest)
				if err := renderPage(w, "create_post.html", data); err != nil {
					return
				}
				return
			}
			postID, err := store.CreatePost(user.ID, title, content)
			if err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not create the post.")
				return
			}
			for _, cidStr := range categoryIDs {
				cid, err := strconv.ParseInt(cidStr, 10, 64)
				if err != nil {
					RenderError(w, http.StatusBadRequest, "Invalid category.")
					return
				}
				if err := store.AddPostCategory(postID, cid); err != nil {
					RenderError(w, http.StatusInternalServerError, "Could not save the post category.")
					return
				}
			}
			http.Redirect(w, r, "/post?id="+strconv.FormatInt(postID, 10), http.StatusSeeOther)
			return
		}

		RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
	}
}

type viewComment struct {
	ID           int64
	Author       string
	Content      string
	CreatedAt    time.Time
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
		user, _ := middleware.GetUserFromRequest(store, r)

		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			RenderError(w, http.StatusBadRequest, "Invalid post ID.")
			return
		}

		post, err := store.GetPostByID(id)
		if err != nil {
			RenderError(w, http.StatusNotFound, "Post not found.")
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
			commentAuthor, err := store.GetUserByID(c.AuthorID)
			if err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not load comment author.")
				return
			}
			likes, err := store.CountCommentLikes(c.ID)
			if err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not load reactions.")
				return
			}
			dislikes, err := store.CountCommentDislikes(c.ID)
			if err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not load reactions.")
				return
			}
			userReaction := ""
			if user != nil {
				userReaction, err = store.GetCommentReaction(c.ID, user.ID)
				if err != nil {
					RenderError(w, http.StatusInternalServerError, "Could not load reactions.")
					return
				}
			}
			vcMap[c.ID] = viewComment{
				ID:           c.ID,
				Author:       commentAuthor.Username,
				Content:      c.Content,
				CreatedAt:    c.CreatedAt,
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
