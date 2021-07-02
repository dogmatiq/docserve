package templates

import (
	"embed"
	"html/template"

	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin/render"
)

func NewRenderer() render.HTMLRender {
	mt := multitemplate.New()

	add(&mt, "applications.html")
	add(&mt, "handlers.html")
	add(&mt, "messages.html")

	return mt
}

//go:embed *.html
var fs embed.FS

func add(mt *multitemplate.Render, name string) {
	tmpl := template.Must(
		template.ParseFS(fs, "layout.html", name),
	)

	mt.Add(name, tmpl)
}
