package handlers

import (
	"forum/db"
	"forum/middleware"
	"net/http"
	"strconv"
	"strings"
)

func CreateCommentHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := middleware.GetUserFromRequest(store, r)

		postIDStr := r.FormValue("post_id")
		content := strings.TrimSpace(r.FormValue("content"))
		if content == "" {
			RenderError(w, http.StatusBadRequest, "Comment text is required.")
			return
		}
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			RenderError(w, http.StatusBadRequest, "Invalid post ID.")
			return
		}
		var parentID *int64
		if pidStr := r.FormValue("parent_id"); pidStr != "" {
			pid, err := strconv.ParseInt(pidStr, 10, 64)
			if err != nil {
				RenderError(w, http.StatusBadRequest, "Invalid parent ID.")
				return
			}
			parentID = &pid
		}
		newID, err := store.CreateCommentWithParent(postID, user.ID, parentID, content)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not create the comment.")
			return
		}
		// For replies scroll to the parent comment; for top-level scroll to the new comment.
		scrollID := newID
		if parentID != nil {
			scrollID = *parentID
		}
		http.Redirect(w, r, "/post?id="+strconv.FormatInt(postID, 10)+"&scroll=comment-"+strconv.FormatInt(scrollID, 10), http.StatusSeeOther)
	}
}
