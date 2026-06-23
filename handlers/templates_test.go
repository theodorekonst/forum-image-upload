package handlers

import "testing"

func TestPageTemplatesParse(t *testing.T) {
	pages := []string{
		"home.html",
		"login.html",
		"register.html",
		"post.html",
		"create_post.html",
	}

	for _, page := range pages {
		if _, err := parsePageTemplate(page); err != nil {
			t.Fatalf("parse %s: %v", page, err)
		}
	}
}
