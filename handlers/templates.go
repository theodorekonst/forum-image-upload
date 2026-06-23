package handlers

import (
	"bytes"
	"html/template"
	"net/http"
)

func renderPage(w http.ResponseWriter, filename string, data any) error {
	tmpl, err := parsePageTemplate(filename)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	_, err = body.WriteTo(w)
	return err
}

func parsePageTemplate(filename string) (*template.Template, error) {
	tmpl, err := template.ParseFiles("templates/" + filename)
	if err == nil {
		return tmpl, nil
	}

	// Tests run from handlers/, while the server runs from the project root.
	return template.ParseFiles("../templates/" + filename)
}
