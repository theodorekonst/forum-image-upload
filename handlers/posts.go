package handlers

import (
	"forum/db"
	"forum/middleware"
	"html/template"
	"net/http"
	"strconv"
)

var postTmpl = template.Must(template.ParseFiles("templates/post.html"))
var createPostTmpl = template.Must(template.ParseFiles("templates/createPost.html"))

func CreatePostHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.GetUserFromRequest(store, r)
		if r.Method == http.MethodGet {
			categories, err := store.ListCategories()
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			createPostTmpl.Execute(w, map[string]any{
				"User":       user,
				"Categories": categories,
			})
			return

		}
		if r.Method == http.MethodPost {
			title := r.FormValue("title")
			content := r.FormValue("content")
			categoryIDs := r.Form["categories"]
			if title == "" || content == "" {
				http.Error(w, "title and content are required", http.StatusBadRequest)
				return
			}
			postID, err := store.CreatePost(user.ID, title, content)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			for _, cidStr := range categoryIDs {
				cid, err := strconv.ParseInt(cidStr, 10, 64)
				if err == nil {
					store.AddPostCategory(postID, cid)
				}
			}
			http.Redirect(w, r, "/post?id="+strconv.FormatInt(postID, 10), http.StatusSeeOther)

		}

	}
}

func ViewPostHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.GetUserFromRequest(store, r)

		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid post id", http.StatusBadRequest)
			return
		}

		post, err := store.GetPostByID(id)
		if err != nil {
			http.Error(w, "post not found", http.StatusNotFound)
			return
		}

		comments, err := store.ListCommentsByPost(id)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		postTmpl.Execute(w, map[string]any{
			"User":     user,
			"Post":     post,
			"Comments": comments,
		})
	}
}
