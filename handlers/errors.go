package handlers

import (
	"bytes"
	"html/template"
	"net/http"
)

type errorPageData struct {
	StatusCode int
	Title      string
	Message    string
}

// RenderError writes a status code and renders the shared error page.
func RenderError(w http.ResponseWriter, statusCode int, message string) {
	if message == "" {
		message = http.StatusText(statusCode)
	}

	data := errorPageData{
		StatusCode: statusCode,
		Title:      http.StatusText(statusCode),
		Message:    message,
	}

	tmpl, err := parseErrorTemplate()
	if err != nil {
		http.Error(w, message, statusCode)
		return
	}

	// Render before writing the status so we can still fall back cleanly.
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		http.Error(w, message, statusCode)
		return
	}

	w.WriteHeader(statusCode)
	body.WriteTo(w)
}

func parseErrorTemplate() (*template.Template, error) {
	tmpl, err := template.ParseFiles("templates/error.html")
	if err == nil {
		return tmpl, nil
	}

	// Tests may run from the handlers package directory.
	return template.ParseFiles("../templates/error.html")
}
