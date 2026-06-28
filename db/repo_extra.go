package db

import (
	"database/sql"
	"errors"
)

// DeleteSession removes a session by its token and returns ErrNotFound if nothing was deleted.
func (s *Store) DeleteSession(token string) error {

	// Delete session row
	result, err := s.DB.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	if err != nil {
		return err
	}

	// Check if a row was actually removed
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteSessionsByUserID removes all sessions belonging to a specific user.
func (s *Store) DeleteSessionsByUserID(userID int64) error {

	// Delete all sessions for this user
	_, err := s.DB.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

func (s *Store) GetPostReaction(postID, userID int64) (string, error) {
	var reactionType string
	err := s.DB.QueryRow(`SELECT reaction_type FROM post_likes WHERE post_id = ? AND user_id = ?`, postID, userID).Scan(&reactionType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return reactionType, nil
}

func (s *Store) SetPostReaction(postID, userID int64, reactionType string) error {
	_, err := s.DB.Exec(
		`INSERT INTO post_likes (post_id, user_id, reaction_type) VALUES (?, ?, ?)
		 ON CONFLICT(post_id, user_id) DO UPDATE SET reaction_type = excluded.reaction_type`,
		postID, userID, reactionType,
	)
	return err
}

func (s *Store) RemovePostReaction(postID, userID int64) error {
	_, err := s.DB.Exec(`DELETE FROM post_likes WHERE post_id = ? AND user_id = ?`, postID, userID)
	return err
}

func (s *Store) GetCommentReaction(commentID, userID int64) (string, error) {
	var reactionType string
	err := s.DB.QueryRow(`SELECT reaction_type FROM comment_likes WHERE comment_id = ? AND user_id = ?`, commentID, userID).Scan(&reactionType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return reactionType, nil
}

func (s *Store) SetCommentReaction(commentID, userID int64, reactionType string) error {
	_, err := s.DB.Exec(
		`INSERT INTO comment_likes (comment_id, user_id, reaction_type) VALUES (?, ?, ?)
		 ON CONFLICT(comment_id, user_id) DO UPDATE SET reaction_type = excluded.reaction_type`,
		commentID, userID, reactionType,
	)
	return err
}

func (s *Store) RemoveCommentReaction(commentID, userID int64) error {
	_, err := s.DB.Exec(`DELETE FROM comment_likes WHERE comment_id = ? AND user_id = ?`, commentID, userID)
	return err
}

func (s *Store) CountPostLikes(postID int64) (int64, error) {
	var count int64
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND reaction_type = 'like'`, postID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) CountPostDislikes(postID int64) (int64, error) {
	var count int64
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND reaction_type = 'dislike'`, postID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) CountCommentLikes(commentID int64) (int64, error) {
	var count int64
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND reaction_type = 'like'`, commentID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) CountCommentDislikes(commentID int64) (int64, error) {
	var count int64
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND reaction_type = 'dislike'`, commentID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) ListPostsByCategory(categoryID int64, limit, offset int) ([]*Post, error) {
	rows, err := s.DB.Query(`SELECT p.id, p.author_id, p.title, p.content, p.created_at FROM posts p JOIN post_categories pc ON p.id = pc.post_id WHERE pc.category_id = ? ORDER BY p.created_at DESC LIMIT ? OFFSET ?`, categoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, rows.Err()
}

func (s *Store) ListCategoriesByPost(postID int64) ([]string, error) {
	rows, err := s.DB.Query(`SELECT c.name FROM categories c JOIN post_categories pc ON c.id = pc.category_id WHERE pc.post_id = ? ORDER BY c.name`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (s *Store) ListPostsByAuthor(authorID int64, limit, offset int) ([]*Post, error) {
	rows, err := s.DB.Query(`SELECT id, author_id, title, content, created_at FROM posts WHERE author_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, authorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, rows.Err()
}

func (s *Store) ListLikedPosts(userID int64, limit, offset int) ([]*Post, error) {
	rows, err := s.DB.Query(`SELECT p.id, p.author_id, p.title, p.content, p.created_at FROM posts p JOIN post_likes pl ON p.id = pl.post_id WHERE pl.user_id = ? ORDER BY p.created_at DESC LIMIT ? OFFSET ?`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, rows.Err()
}
