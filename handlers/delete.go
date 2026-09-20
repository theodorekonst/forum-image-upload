package handlers

import (
	"errors"
	"forum/db"
	"forum/middleware"
	"net/http"
	"strconv"
)

func DeletePostHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
			return
		}

		user, err := middleware.GetUserFromRequest(store, r)
		if err != nil || user == nil {
			RenderError(w, http.StatusUnauthorized, "You must be logged in to delete a post.")
			return
		}

		postID, err := strconv.ParseInt(r.FormValue("post_id"), 10, 64)
		if err != nil {
			RenderError(w, http.StatusBadRequest, "Invalid post ID.")
			return
		}

		post, err := store.GetPostByID(postID)
		if errors.Is(err, db.ErrNotFound) {
			RenderError(w, http.StatusNotFound, "Post not found.")
			return
		}
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not load the post.")
			return
		}

		if post.AuthorID != user.ID {
			RenderError(w, http.StatusForbidden, "You can only delete your own posts.")
			return
		}

		if err := store.DeletePost(postID); err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not delete the post.")
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func DeleteCommentHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
			return
		}

		user, err := middleware.GetUserFromRequest(store, r)
		if err != nil || user == nil {
			RenderError(w, http.StatusUnauthorized, "You must be logged in to delete a comment.")
			return
		}

		commentID, err := strconv.ParseInt(r.FormValue("comment_id"), 10, 64)
		if err != nil {
			RenderError(w, http.StatusBadRequest, "Invalid comment ID.")
			return
		}

		comment, err := store.GetCommentByID(commentID)
		if errors.Is(err, db.ErrNotFound) {
			RenderError(w, http.StatusNotFound, "Comment not found.")
			return
		}
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not load the comment.")
			return
		}

		if comment.IsDeleted || comment.AuthorID != user.ID {
			RenderError(w, http.StatusForbidden, "You can only delete your own comments.")
			return
		}

		if _, err := store.DeleteComment(commentID); err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not delete the comment.")
			return
		}

		http.Redirect(w, r, "/post?id="+strconv.FormatInt(comment.PostID, 10), http.StatusSeeOther)
	}
}
