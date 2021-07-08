package templates

import (
	"embed"
	"html/template"

	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin/render"
)

func NewRenderer() render.HTMLRender {
	mt := multitemplate.New()

	add(&mt, "error.html")

	add(&mt, "application-list.html")
	add(&mt, "application-view.html")
	add(&mt, "handler-list.html")
	add(&mt, "handler-view.html")
	add(&mt, "message-list.html")
	add(&mt, "message-view.html")

	return mt
}

//go:embed *.html
var fs embed.FS

var funcMap = template.FuncMap{}

// add adds a template to the multi-template render.
//
// note that we can not use template.ParseFS() because it does not allow us to
// specify the function map, which must be done before parsing.
func add(mt *multitemplate.Render, name string) {
	tmpl := template.
		New("").
		Funcs(funcMap)

	for _, file := range []string{
		"layout.html",
		name,
	} {
		content, err := fs.ReadFile(file)
		if err != nil {
			panic(err)
		}

		tmpl = template.Must(
			tmpl.Parse(string(content)),
		)
	}

	mt.Add(name, tmpl)
}
