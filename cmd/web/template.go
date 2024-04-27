package main

import (
	"html/template"
	"io/fs"
	"path/filepath"

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
type templateData struct {
	WorkerRegistry  *data.RegistryWorker
	WorkersRegistry []*data.RegistryWorker
	Form            any
	Flash           string //message to be displayed to the user
	IsAuthenticated bool   // authenticated user
}

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
