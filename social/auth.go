package main

import (
	"context"
	"github.com/chocosin/otus-hl/social/model"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"time"
)

const CookieName = "auth_token"

type contextKeyAuth = int

const UserKey contextKeyAuth = 0
const TokenKey contextKeyAuth = 1

func (app *App) checkAuthedAndRedirect(authed bool, redirectTo string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r.Context())
			userAuthed := user != nil
			if userAuthed == authed {
				redirect(w, r, redirectTo)
				return
			}
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(f)
	}
}

func SetAuthCookie(w http.ResponseWriter, token uuid.UUID) {
	http.SetCookie(w, &http.Cookie{
		Name:    CookieName,
		Value:   token.String(),
		Expires: time.Now().Add(time.Hour * 24 * 30),
	})
}

func RemoveAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   CookieName,
		MaxAge: -1,
	})
}

func (app *App) auth(h http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieName)
		if err != nil {
			if err != http.ErrNoCookie {
				app.logger.Err(err).Msg("failed to get auth cookie")
			}
			h.ServeHTTP(w, r)
			return
		}
		token, err := uuid.FromString(cookie.Value)
		if err != nil {
			app.logger.Err(err).
				Str("cookie", cookie.Value).
				Msg("failed to parse cookie")
			h.ServeHTTP(w, r)
			return
		}
		user, err := app.storage.GetUserByToken(token)
		if err != nil {
			app.logger.Err(err).
				Str("token", token.String()).
				Msg("failed to retrieve user by token")
			h.ServeHTTP(w, r)
			return
		}
		if user != nil {
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserKey, user)
			ctx = context.WithValue(ctx, TokenKey, token)
			r = r.WithContext(ctx)
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}

func GetUser(ctx context.Context) *model.User {
	value := ctx.Value(UserKey)
	if usr, ok := value.(*model.User); ok {
		return usr
	}
	return nil
}

func GetToken(ctx context.Context) uuid.UUID {
	value := ctx.Value(TokenKey)
	if token, ok := value.(uuid.UUID); ok {
		return token
	}
	return uuid.Nil
}
