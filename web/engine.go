package web

import (
	"fmt"
	"net/http"

	"github.com/dogmatiq/docserve/web/templates"
	"github.com/gin-gonic/gin"
)

type Handler interface {
	Route() (string, string)
	ServeHTTP(*gin.Context) error
}

func NewRouter(handlers ...Handler) http.Handler {
	engine := gin.New()
	engine.HTMLRender = templates.NewRenderer()

	for _, h := range handlers {
		method, path := h.Route()

		engine.Handle(
			method,
			path,
			wrap(h.ServeHTTP),
		)
	}

	return engine
}

func wrap(handle func(*gin.Context) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := handle(ctx); err != nil {
			err = ctx.AbortWithError(
				http.StatusInternalServerError,
				err,
			)
			ctx.Writer.WriteString("Internal server error") // nolint:errcheck
			fmt.Println(err)                                // TODO
		}
	}
}
