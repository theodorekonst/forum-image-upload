package handlers

import (
	"errors"
	"forum/db"
	"forum/middleware"
	"net/http"
	"strings"
	"time"
)

const homePageSize = 50

type homePageData struct {
	User           *db.User
	Posts          []homePost
	Categories     []*db.Category
	ActiveCategory string
	ActiveFilter   string
}

type homePost struct {
	ID         int64
	Title      string
	Author     string
	CreatedAt  time.Time
	Excerpt    string
	Categories []string
	Likes      int64
	Dislikes   int64
}

// HomeHandler renders the public post feed and applies optional filters.
func HomeHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			RenderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
			return
		}

		user, _ := middleware.GetUserFromRequest(store, r)
		categories, err := store.ListCategories()
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not load categories.")
			return
		}

		posts, activeCategory, activeFilter, err := loadHomePosts(store, r, user)
		if err != nil {
			if errors.Is(err, errLoginRequired) {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			RenderError(w, http.StatusInternalServerError, "Could not load posts.")
			return
		}

		viewPosts, err := buildHomePosts(store, posts)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not prepare posts.")
			return
		}

		if err := renderPage(w, "home.html", homePageData{
			User:           user,
			Posts:          viewPosts,
			Categories:     categories,
			ActiveCategory: activeCategory,
			ActiveFilter:   activeFilter,
		}); err != nil {
			RenderError(w, http.StatusInternalServerError, "Could not render the home page.")
		}
	}
}

var errLoginRequired = errors.New("login required")

func loadHomePosts(store *db.Store, r *http.Request, user *db.User) ([]*db.Post, string, string, error) {
	categoryName := strings.TrimSpace(r.URL.Query().Get("category"))
	filter := r.URL.Query().Get("filter")

	if categoryName != "" {
		category, err := store.GetCategoryByName(categoryName)
		if errors.Is(err, db.ErrNotFound) {
			return []*db.Post{}, categoryName, "", nil
		}
		if err != nil {
			return nil, "", "", err
		}
		posts, err := store.ListPostsByCategory(category.ID, homePageSize, 0)
		return posts, category.Name, "", err
	}

	switch filter {
	case "", "all":
		posts, err := store.ListPosts(homePageSize, 0)
		return posts, "", "all", err
	case "mine":
		if user == nil {
			return nil, "", "", errLoginRequired
		}
		posts, err := store.ListPostsByAuthor(user.ID, homePageSize, 0)
		return posts, "", "mine", err
	case "liked":
		if user == nil {
			return nil, "", "", errLoginRequired
		}
		posts, err := store.ListLikedPosts(user.ID, homePageSize, 0)
		return posts, "", "liked", err
	default:
		return nil, "", "", errors.New("unknown filter")
	}
}

func buildHomePosts(store *db.Store, posts []*db.Post) ([]homePost, error) {
	viewPosts := make([]homePost, 0, len(posts))
	for _, post := range posts {
		author, err := store.GetUserByID(post.AuthorID)
		if err != nil {
			return nil, err
		}
		categories, err := store.ListCategoriesByPost(post.ID)
		if err != nil {
			return nil, err
		}
		likes, err := store.CountPostLikes(post.ID)
		if err != nil {
			return nil, err
		}

		viewPosts = append(viewPosts, homePost{
			ID:         post.ID,
			Title:      post.Title,
			Author:     author.Username,
			CreatedAt:  post.CreatedAt,
			Excerpt:    excerpt(post.Content),
			Categories: categories,
			Likes:      likes,
			Dislikes:   0,
		})
	}
	return viewPosts, nil
}

func excerpt(content string) string {
	const maxLength = 180
	content = strings.TrimSpace(content)
	runes := []rune(content)
	if len(runes) <= maxLength {
		return content
	}
	return string(runes[:maxLength]) + "..."
}
