package web

import (
	"fmt"
	"net/http"

	"github.com/dogmatiq/docserve/web/templates"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type Handler interface {
	Route() (string, string)
	ServeHTTP(*gin.Context) error
}

func NewRouter(
	c *oauth2.Config,
	handlers ...Handler,
) http.Handler {
	router := gin.New()
	router.HTMLRender = templates.NewRenderer()

	router.Use(gin.Recovery())

	router.GET(
		"/github/auth",
		handleOAuthCallback(c),
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

func adaptHandler(h Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := h.ServeHTTP(ctx); err != nil {
			fmt.Println(err) // TODO
			renderError(ctx, http.StatusInternalServerError)
			return
		}
	}
}

func renderError(ctx *gin.Context, code int) {
	var tctx = struct {
		templates.Context
		StatusCode int
		Message    string
	}{
		Context: templates.Context{
			Title: http.StatusText(code),
		},
		StatusCode: code,
	}

	switch code {
	case http.StatusNotFound:
		tctx.Message = "The content you have requested can not be found."
	case http.StatusUnauthorized:
		tctx.Message = "You do not have permission to view this context."
	case http.StatusInternalServerError:
		tctx.Message = "An unexpected error occurred on the server."
	}

	ctx.HTML(http.StatusInternalServerError, "error.html", tctx)
}
