package handlers

import (
	"forum/db"
	"forum/middleware"
	"net/http"
	"strconv"
)

func PostReactHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
			return
		}

		user, _ := middleware.GetUserFromRequest(store, r)

		postIDStr := r.FormValue("post_id")
		reactionType := r.FormValue("reaction_type")

		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			RenderError(w, http.StatusBadRequest, "Invalid post ID.")
			return
		}

		if reactionType != "like" && reactionType != "dislike" {
			RenderError(w, http.StatusBadRequest, "Invalid reaction type.")
			return
		}

		current, err := store.GetPostReaction(postID, user.ID)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not check reaction.")
			return
		}

		if current == reactionType {
			if err := store.RemovePostReaction(postID, user.ID); err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not remove reaction.")
				return
			}
		} else {
			if err := store.SetPostReaction(postID, user.ID, reactionType); err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not set reaction.")
				return
			}
		}

		redirect := r.Referer()
		if redirect == "" {
			redirect = "/post?id=" + postIDStr
		}
		http.Redirect(w, r, redirect, http.StatusSeeOther)
	}
}

func CommentReactHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
			return
		}

		user, _ := middleware.GetUserFromRequest(store, r)

		commentIDStr := r.FormValue("comment_id")
		postIDStr := r.FormValue("post_id")
		reactionType := r.FormValue("reaction_type")

		commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
		if err != nil {
			RenderError(w, http.StatusBadRequest, "Invalid comment ID.")
			return
		}
		comment, err := store.GetCommentByID(commentID)
		if err != nil || comment.IsDeleted {
			RenderError(w, http.StatusNotFound, "Comment not found.")
			return
		}

		if reactionType != "like" && reactionType != "dislike" {
			RenderError(w, http.StatusBadRequest, "Invalid reaction type.")
			return
		}

		current, err := store.GetCommentReaction(commentID, user.ID)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not check reaction.")
			return
		}

		if current == reactionType {
			if err := store.RemoveCommentReaction(commentID, user.ID); err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not remove reaction.")
				return
			}
		} else {
			if err := store.SetCommentReaction(commentID, user.ID, reactionType); err != nil {
				RenderError(w, http.StatusInternalServerError, "Could not set reaction.")
				return
			}
		}

		http.Redirect(w, r, "/post?id="+postIDStr, http.StatusSeeOther)
	}
}
