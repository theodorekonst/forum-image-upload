package handlers

import (
	"forum/db"
	"forum/middleware"
	"net/http"
	"strconv"
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
			title := r.FormValue("title")
			content := r.FormValue("content")
			categoryIDs := r.Form["categories"]
			if title == "" || content == "" {
				RenderError(w, http.StatusBadRequest, "Title and content are required.")
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
}

type postPageData struct {
	User             *db.User
	Post             *db.Post
	PostAuthor       string
	PostLikes        int64
	PostDislikes     int64
	PostUserReaction string
	Comments         []viewComment
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

		comments := make([]viewComment, 0, len(rawComments))
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
			comments = append(comments, viewComment{
				ID:           c.ID,
				Author:       commentAuthor.Username,
				Content:      c.Content,
				CreatedAt:    c.CreatedAt,
				Likes:        likes,
				Dislikes:     dislikes,
				UserReaction: userReaction,
			})
		}

		if err := renderPage(w, "post.html", postPageData{
			User:             user,
			Post:             post,
			PostAuthor:       postAuthor.Username,
			PostLikes:        postLikes,
			PostDislikes:     postDislikes,
			PostUserReaction: postUserReaction,
			Comments:         comments,
		}); err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not render the post page.")
		}
	}
}
