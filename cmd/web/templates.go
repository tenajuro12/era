// cmd/web/routes.go

package main

import (
	"Movies/pkg/forms"
	"Movies/pkg/models"
	"html/template"
	"path/filepath"
	"time"
)

type templateData struct {
	CSRFToken string

	CurrentYear     int
	Flash           string
	Form            *forms.Form
	IsAuthenticated bool

	Movies  *models.Movies
	Movies2 []*models.Movies
	User    *models.User
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 15:04:59")
}
func humanTime(t time.Time) string {
	return t.Format("15:04:59")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
	"humanTime": humanTime,
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {

		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
