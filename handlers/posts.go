package handlers

import (
	"forum/db"
	"forum/middleware"
	"net/http"
	"strconv"
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

		comments, err := store.ListCommentsByPost(id)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not load comments.")
			return
		}

		if err := renderPage(w, "post.html", map[string]any{
			"User":     user,
			"Post":     post,
			"Comments": comments,
		}); err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not render the post page.")
		}
	}
}
