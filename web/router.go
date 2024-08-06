package web

import (
	"crypto/rsa"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/dogmatiq/browser/analyzer"
	"github.com/dogmatiq/browser/githubx"
	"github.com/dogmatiq/browser/web/components"
	"github.com/dogmatiq/browser/web/pages/applications"
	"github.com/dogmatiq/browser/web/pages/handlers"
	"github.com/dogmatiq/browser/web/pages/messages"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v38/github"
)

// Handler is an interface for handlers that render human-readable pages.
type Handler interface {
	Route() (string, string)
	Template() string
	ActiveMenuItem() components.MenuItem
	View(ctx *gin.Context) (string, interface{}, error)
}

// NewRouter returns an http.Handler that routes requests to the appropriate
// handler.
func NewRouter(
	version string,
	c *githubx.Connector,
	o *analyzer.Orchestrator,
	key *rsa.PrivateKey,
	hookSecret string,
	db *sql.DB,
) http.Handler {
	engine := gin.New()

	engine.Use(gin.Recovery())
	engine.Use(gin.LoggerWithConfig(
		gin.LoggerConfig{
			Skip: func(ctx *gin.Context) bool {
				return ctx.Request.URL.Path == "/health" ||
					strings.HasPrefix(ctx.Request.URL.Path, "/assets/")
			},
		},
	))

	engine.HTMLRender = pageTemplates

	engine.GET("/health", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "OK")
	})

	assets := http.FileServer(
		http.FS(assetsFS),
	)

	assetsHandler := func(ctx *gin.Context) {
		assets.ServeHTTP(ctx.Writer, ctx.Request)
	}

	engine.GET("/assets/*path", assetsHandler)
	engine.HEAD("/assets/*path", assetsHandler)

	engine.GET(
		"/github/auth",
		handleOAuthCallback(version, c.OAuthConfig, &key.PublicKey),
	)

	engine.POST(
		"/github/hook",
		handleGitHubHook(version, hookSecret, o),
	)

	engine.GET(
		"/search/items.json",
		searchItems(version, db),
	)

	auth := requireOAuth(version, c, key)

	handlers := [...]Handler{
		&applications.ListHandler{DB: db},
		&applications.DetailsHandler{DB: db},
		&applications.RelationshipHandler{DB: db},
		&handlers.ListHandler{DB: db},
		&handlers.DetailsHandler{DB: db},
		&messages.ListHandler{DB: db},
		&messages.DetailsHandler{DB: db},
	}

	for _, h := range handlers {
		method, path := h.Route()

		engine.Handle(
			method,
			path,
			auth,
			adaptHandler(version, h),
		)
	}

	engine.NoRoute(
		func(ctx *gin.Context) {
			renderError(ctx, version, http.StatusNotFound)
		},
	)

	return engine
}

type templateContext struct {
	Title          string
	ActiveMenuItem components.MenuItem
	User           *github.User
	Version        string
	View           interface{}
}

func adaptHandler(version string, h Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		title, view, err := h.View(ctx)
		if err != nil {
			fmt.Println("unable to generate view:", err) // TODO
			renderError(ctx, version, http.StatusInternalServerError)
			return
		}

		if ctx.IsAborted() {
			code := ctx.Writer.Status()
			if code == http.StatusOK {
				panic("handler aborted execution with OK status")
			}

			renderError(
				ctx,
				version,
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
				Version:        version,
				View:           view,
			},
		)
	}
}

func renderError(ctx *gin.Context, version string, code int) {
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
			Title:   http.StatusText(code),
			User:    u,
			Version: version,
			View:    view,
		},
	)

	if ctx.Writer.Written() {
		renderer.Render(ctx.Writer) // nolint:errcheck
		return
	}

	ctx.Render(code, renderer)
}
