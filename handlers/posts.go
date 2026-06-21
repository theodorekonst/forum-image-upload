package handlers

import (
	"forum/db"
	"forum/middleware"
	"html/template"
	"net/http"
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

		}

	}
}
