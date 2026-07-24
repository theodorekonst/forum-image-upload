package handlers

import (
	"errors"
	"forum/db"
	"forum/middleware"
	"net/http"
	"strconv"
	"strings"
)

func CreateCommentHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
			return
		}
		user, _ := middleware.GetUserFromRequest(store, r)

		postIDStr := r.FormValue("post_id")
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil || postID <= 0 {
			RenderError(w, http.StatusBadRequest, "Invalid post ID.")
			return
		}
		content := strings.TrimSpace(r.FormValue("content"))
		if content == "" {
			http.Redirect(w, r, "/post?id="+strconv.FormatInt(postID, 10)+"&comment_error=required&scroll=comment-form", http.StatusSeeOther)
			return
		}
		if _, err := store.GetPostByID(postID); errors.Is(err, db.ErrNotFound) {
			RenderError(w, http.StatusNotFound, "Post not found.")
			return
		} else if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not load the post.")
			return
		}
		var parentID *int64
		if pidStr := r.FormValue("parent_id"); pidStr != "" {
			pid, err := strconv.ParseInt(pidStr, 10, 64)
			if err != nil {
				RenderError(w, http.StatusBadRequest, "Invalid parent ID.")
				return
			}
			parent, err := store.GetCommentByID(pid)
			if errors.Is(err, db.ErrNotFound) {
				RenderError(w, http.StatusBadRequest, "Invalid parent comment.")
				return
			}
			if err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not load the parent comment.")
				return
			}
			if parent.IsDeleted || parent.PostID != postID {
				RenderError(w, http.StatusBadRequest, "Invalid parent comment.")
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
