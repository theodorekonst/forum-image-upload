package db

import (
	"database/sql"
	"errors"
	"time"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

type User struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	PasswordHash  string    `json:"password_hash"`
	OAuthProvider string    `json:"oauth_provider"`
	OAuthID       string    `json:"oauth_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type Post struct {
	ID        int64     `json:"id"`
	AuthorID  int64     `json:"author_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	AuthorID  int64     `json:"author_id"`
	Content   string    `json:"content"`
	ParentID  *int64    `json:"parent_id"`
	IsDeleted bool      `json:"is_deleted"`
	CreatedAt time.Time `json:"created_at"`
}

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Session struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

var ErrNotFound = errors.New("record not found")
var ErrCategoryRequired = errors.New("at least one category is required")
var ErrInvalidCategory = errors.New("invalid category")

func (s *Store) CreateUser(email, username, passwordHash string) (int64, error) {
	result, err := s.DB.Exec(
		`INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)`,
		email,
		username,
		passwordHash,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) GetUserByID(id int64) (*User, error) {
	row := s.DB.QueryRow(
		`SELECT id, email, username, password_hash, oauth_provider, oauth_id, created_at FROM users WHERE id = ?`,
		id,
	)
	return scanUser(row)
}

func (s *Store) GetUserByEmail(email string) (*User, error) {
	row := s.DB.QueryRow(
		`SELECT id, email, username, password_hash, oauth_provider, oauth_id, created_at FROM users WHERE email = ?`,
		email,
	)
	return scanUser(row)
}

func (s *Store) GetUserByUsername(username string) (*User, error) {
	row := s.DB.QueryRow(
		`SELECT id, email, username, password_hash, oauth_provider, oauth_id, created_at FROM users WHERE username = ?`,
		username,
	)
	return scanUser(row)
}

// GetUserByOAuth returns a user by OAuth provider and ID or ErrNotFound if it does not exist.
func (s *Store) GetUserByOAuth(provider, oauthID string) (*User, error) {
	// Query user by OAuth provider and ID
	row := s.DB.QueryRow(
		`SELECT id, email, username, password_hash, oauth_provider, oauth_id, created_at FROM users WHERE oauth_provider = ? AND oauth_id = ?`,
		provider, oauthID,
	)
	// Scan into struct
	return scanUser(row)
}

// CreateOAuthUser creates a new user record for an OAuth account.
func (s *Store) CreateOAuthUser(email, username, provider, oauthID string) (int64, error) {
	// Insert new OAuth user row
	result, err := s.DB.Exec(
		`INSERT INTO users (email, username, password_hash, oauth_provider, oauth_id)
         VALUES (?, ?, '', ?, ?)`,
		email, username, provider, oauthID,
	)
	if err != nil {
		return 0, err
	}
	// Return new user ID
	return result.LastInsertId()
}

func scanUser(row scanner) (*User, error) {
	var user User
	if err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.OAuthProvider,
		&user.OAuthID,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func (s *Store) CreatePost(authorID int64, title, content string) (int64, error) {
	result, err := s.DB.Exec(
		`INSERT INTO posts (author_id, title, content) VALUES (?, ?, ?)`,
		authorID,
		title,
		content,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) GetPostByID(id int64) (*Post, error) {
	row := s.DB.QueryRow(`SELECT id, author_id, title, content, created_at FROM posts WHERE id = ?`, id)
	var post Post
	if err := row.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (s *Store) DeletePost(id int64) error {
	_, err := s.DB.Exec(
		`DELETE FROM posts WHERE id = ?`,
		id,
	)
	return err
}

func (s *Store) ListPosts(limit, offset int) ([]*Post, error) {
	rows, err := s.DB.Query(`SELECT id, author_id, title, content, created_at FROM posts ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
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

func (s *Store) CreateComment(postID, authorID int64, content string) (int64, error) {
	result, err := s.DB.Exec(
		`INSERT INTO comments (post_id, author_id, content) VALUES (?, ?, ?)`,
		postID,
		authorID,
		content,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) ListCommentsByPost(postID int64) ([]*Comment, error) {
	rows, err := s.DB.Query(`SELECT id, post_id, author_id, content, parent_id, is_deleted, created_at FROM comments WHERE post_id = ? ORDER BY created_at`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.AuthorID, &comment.Content, &comment.ParentID, &comment.IsDeleted, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}
	return comments, rows.Err()
}

func (s *Store) CreateCategory(name string) (int64, error) {
	result, err := s.DB.Exec(`INSERT INTO categories (name) VALUES (?)`, name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) GetCategoryByName(name string) (*Category, error) {
	row := s.DB.QueryRow(`SELECT id, name, created_at FROM categories WHERE name = ?`, name)
	var category Category
	if err := row.Scan(&category.ID, &category.Name, &category.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (s *Store) ListCategories() ([]*Category, error) {
	rows, err := s.DB.Query(`SELECT id, name, created_at FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, &category)
	}
	return categories, rows.Err()
}

func (s *Store) AddPostCategory(postID, categoryID int64) error {
	_, err := s.DB.Exec(`INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES (?, ?)`, postID, categoryID)
	return err
}

// CreateSession creates a new session for a user with a unique token and expiration time.
func (s *Store) CreateSession(userID int64, token string, expiresAt time.Time) (int64, error) {

	// Get current time in local timezone
	now := time.Now()

	// Insert new session row
	result, err := s.DB.Exec(
		`INSERT INTO sessions (user_id, token, expires_at, created_at) VALUES (?, ?, ?, ?)`,
		userID, token, expiresAt, now,
	)
	if err != nil {
		return 0, err
	}

	// Return new session ID
	return result.LastInsertId()
}

// GetSession returns a session by its token or ErrNotFound if it does not exist.
func (s *Store) GetSession(token string) (*Session, error) {

	// Query session by token
	row := s.DB.QueryRow(
		`SELECT id, user_id, token, expires_at, created_at FROM sessions WHERE token = ?`,
		token,
	)

	// Scan into struct
	var session Session
	if err := row.Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &session, nil
}
