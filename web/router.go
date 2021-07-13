package web

import (
	"fmt"
	"net/http"

	"github.com/dogmatiq/docserve/githubx"
	"github.com/dogmatiq/docserve/web/components"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v35/github"
)

type Handler interface {
	Route() (string, string)
	Template() string
	ActiveMenuItem() components.MenuItem
	View(ctx *gin.Context) (string, interface{}, error)
}

func NewRouter(
	c *githubx.Connector,
	handlers ...Handler,
) http.Handler {
	router := gin.New()
	router.HTMLRender = pageTemplates

	router.Use(gin.Recovery())

	router.GET(
		"/github/auth",
		handleOAuthCallback(c.OAuthConfig),
	)

	auth := requireOAuth(c)

	for _, h := range handlers {
		method, path := h.Route()

		router.Handle(
			method,
			path,
			auth,
			adaptHandler(h),
		)
	}

	router.NoRoute(
		func(ctx *gin.Context) {
			renderError(ctx, http.StatusNotFound)
		},
	)

	return router
}

type templateContext struct {
	Title          string
	ActiveMenuItem components.MenuItem
	User           *github.User
	View           interface{}
}

func adaptHandler(h Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		title, view, err := h.View(ctx)
		if err != nil {
			fmt.Println("unable to generate view:", err) // TODO
			renderError(ctx, http.StatusInternalServerError)
			return
		}

		if ctx.IsAborted() {
			code := ctx.Writer.Status()
			if code == http.StatusOK {
				panic("handler aborted execution with OK status")
			}

			renderError(
				ctx,
				ctx.Writer.Status(),
			)

			return
		}

		u, _ := currentUser(ctx)

		ctx.HTML(
			http.StatusOK,
			h.Template(),
			templateContext{
				Title:          title,
				User:           u,
				ActiveMenuItem: h.ActiveMenuItem(),
				View:           view,
			},
		)
	}
}

func renderError(ctx *gin.Context, code int) {
	view := struct {
		StatusText string
		StatusCode int
		Message    string
	}{
		StatusText: http.StatusText(code),
		StatusCode: code,
	}

	switch code {
	case http.StatusNotFound:
		view.Message = "The content you have requested can not be found."
	case http.StatusUnauthorized:
		view.Message = "You do not have permission to view this context."
	case http.StatusInternalServerError:
		view.Message = "An unexpected error occurred on the server."
	}

	u, _ := currentUser(ctx)

	renderer := pageTemplates.Instance(
		"error.html",
		templateContext{
			Title: http.StatusText(code),
			User:  u,
			View:  view,
		},
	)

	if ctx.Writer.Written() {
		renderer.Render(ctx.Writer)
		return
	}

	ctx.Render(code, renderer)
}
