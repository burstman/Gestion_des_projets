package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/burstman/baseRegistry/cmd/web/internal/data"
	"github.com/burstman/baseRegistry/cmd/web/ui"
)

// templateData is a struct that holds data to be passed to HTML templates.
//
// WorkerRegistry is a pointer to a RegistryWorker struct, which likely represents
// a registry of workers.
// WorkersRegistry is a slice of pointers to RegistryWorker structs, which likely
// represents a collection of worker registries.
// Form is an arbitrary data type that likely represents a form or form-related data.
// Flash is a string that likely represents a temporary message or notification to
// be displayed to the user.

type Task struct {
	ID          int
	Description string
	AssignedTo  string
	Status      string
	DueDate     time.Time
}

type Projects struct {
	Name         string
	Description  string
	Status       string
	Deadline     time.Time
	CompleatedAt time.Time
	Comment      map[string]string
	Owner        string
	Participants map[string]string
	Tasks        []*Task
}

type ChatHistory struct {
	ChatUser    string
	ChatTime    string
	ChatMessage string
}


type templateData struct {
	Projects        []*Projects
	ChatHistories   []*ChatHistory
	User            *data.User
	Form            any
	Flash           string //message to be displayed to the user
	IsAuthenticated bool   // authenticated user
}

// newTemplateCache creates a new template cache by parsing all HTML template files
// in the "html/pages/" directory and storing them in a map, keyed by the base name
// of the file. The cache includes the base template "html/base.tmpl.html" and any
// partials in the "html/partials/" directory.
//
// If any errors occur during the parsing process, the function will return an error.
// Otherwise, it will return the populated template cache.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		pattern := []string{
			"html/base.tmpl.html",
			"html/partials/*.html",
			page,
		}
		ts, err := template.New(name).ParseFS(ui.Files, pattern...)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}
	return cache, nil

}
