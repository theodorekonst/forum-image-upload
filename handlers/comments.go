package handlers

import (
	"forum/db"
	"forum/middleware"
	"net/http"
	"strconv"
)

func CreateCommentHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := middleware.GetUserFromRequest(store, r)

		postIDStr := r.FormValue("post_id")
		content := r.FormValue("content")
		if content == "" {
			http.Error(w, "content is required", http.StatusBadRequest)
			return
		}
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid post id", http.StatusBadRequest)
			return
		}
		store.CreateComment(postID, user.ID, content)
		http.Redirect(w, r, "/post?id="+strconv.FormatInt(postID, 10), http.StatusSeeOther)

	}
}
