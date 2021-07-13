package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dogmatiq/docserve/githubx"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
)

func handleOAuthCallback(c *oauth2.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := c.Exchange(ctx, ctx.Query("code"))
		if err != nil {
			fmt.Println("unable to perform oauth exchange:", err) // TODO
			renderError(ctx, http.StatusUnauthorized)
			return
		}

		data, err := json.Marshal(token)
		if err != nil {
			fmt.Println("unable to marshal oauth token:", err) // TODO
			renderError(ctx, http.StatusInternalServerError)
			ctx.Abort()
		}

		ctx.SetCookie(
			"token",
			string(data),
			int(time.Until(token.Expiry).Seconds()),
			"",
			"",
			true,
			true,
		)

		ctx.Redirect(http.StatusTemporaryRedirect, ctx.Query("state"))
	}
}

func requireOAuth(c *githubx.Connector) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		data, err := ctx.Cookie("token")
		if err != nil {
			redirectToLogin(ctx, c.OAuthConfig)
			return
		}

		token := &oauth2.Token{}
		if err := json.Unmarshal([]byte(data), token); err != nil {
			redirectToLogin(ctx, c.OAuthConfig)
			return
		}

		client, err := c.UserClient(ctx, token)
		if err != nil {
			fmt.Println("unable to create user client:", err) // TODO
			renderError(ctx, http.StatusInternalServerError)
			ctx.Abort()
		}

		user, res, err := client.Users.Get(ctx, "")
		if err != nil {
			if res.StatusCode == http.StatusUnauthorized {
				redirectToLogin(ctx, c.OAuthConfig)
				return
			}

			fmt.Println("unable to query connected user:", err) // TODO
			renderError(ctx, http.StatusInternalServerError)
			ctx.Abort()
			return
		}

		ctx.Set("github-user", user)
	}
}

func redirectToLogin(ctx *gin.Context, c *oauth2.Config) {
	authURL := c.AuthCodeURL(ctx.Request.URL.String())
	ctx.Redirect(http.StatusTemporaryRedirect, authURL)
	ctx.Abort()
}

func currentUser(ctx *gin.Context) (*github.User, bool) {
	u, ok := ctx.Value("github-user").(*github.User)
	return u, ok
}
