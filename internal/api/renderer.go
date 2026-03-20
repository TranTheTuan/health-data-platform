package api

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

// TemplateRenderer implements echo.Renderer interface for rendering HTML templates.
type TemplateRenderer struct {
	templates map[string]*template.Template
}

// NewTemplateRenderer creates a new TemplateRenderer parsing each template file individually with the base template.
func NewTemplateRenderer(dir string) (*TemplateRenderer, error) {
	// Parse base.html first
	base, err := template.ParseFiles(dir + "/base.html")
	if err != nil {
		return nil, err
	}

	// Glob the others
	files, err := filepath.Glob(dir + "/*.html")
	if err != nil {
		return nil, err
	}

	renderer := &TemplateRenderer{
		templates: make(map[string]*template.Template),
	}

	for _, f := range files {
		name := filepath.Base(f)
		if name == "base.html" {
			continue
		}

		// Clone base and associate the specific layout
		t, err := base.Clone()
		if err != nil {
			return nil, err
		}

		t, err = t.ParseFiles(f)
		if err != nil {
			return nil, err
		}

		renderer.templates[name] = t
	}

	return renderer, nil
}

// Render executes the matching template with the provided data.
func (r *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	t, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	return t.ExecuteTemplate(w, name, data)
}
