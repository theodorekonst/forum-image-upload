package db

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

func (s *Store) AddLikeToPost(postID, userID int64) (int64, error) {
	result, err := s.DB.Exec(`INSERT OR IGNORE INTO post_likes (post_id, user_id) VALUES (?, ?)`, postID, userID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) AddLikeToComment(commentID, userID int64) (int64, error) {
	result, err := s.DB.Exec(`INSERT OR IGNORE INTO comment_likes (comment_id, user_id) VALUES (?, ?)`, commentID, userID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) CountPostLikes(postID int64) (int64, error) {
	var count int64
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM post_likes WHERE post_id = ?`, postID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) CountCommentLikes(commentID int64) (int64, error) {
	var count int64
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM comment_likes WHERE comment_id = ?`, commentID).Scan(&count); err != nil {
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
