package main

import (
	"errors"
	"fmt"
	"github.com/chocosin/otus-hl/social/model"
	"github.com/chocosin/otus-hl/social/storage"
	"github.com/chocosin/otus-hl/social/templates"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"os"
	"time"
)

type App struct {
	logger    zerolog.Logger
	storage   storage.Storage
	Templates *templates.Templates
}

func main() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	logger := zerolog.New(output).With().
		Timestamp().
		Logger()

	var err error

	templatesDir := os.Getenv("TEMPLATES")
	if templatesDir == "" {
		templatesDir = "./social/templates"
	}
	mysqlConfig := storage.NewMysqlConfig()
	storage.CreateDatabase(mysqlConfig, false)
	storage.Migrate(mysqlConfig)

	mysqlStorage, err := storage.NewMysqlStorage(mysqlConfig)
	if err != nil {
		panic(err)
	}
	app := App{
		logger:  logger,
		storage: mysqlStorage,
	}
	app.Templates, err = templates.NewTemplates(templatesDir)
	if err != nil {
		panic(err)
	}

	root := chi.NewRouter()
	root.Use(middleware.RequestLogger(RequestFormatter{&logger}))
	root.Use(middleware.Timeout(time.Second * 3))
	root.Use(middleware.Recoverer)
	root.Use(app.auth)

	root.With(app.checkAuthedAndRedirect(true, "/me")).
		Get("/", app.indexHandler)

	root.Mount("/signup", app.signupHandler())
	root.Mount("/login", app.loginHandler())
	root.Mount("/user/", app.usersHandler())
	root.Mount("/last", app.lastUsernamesHandler())
	root.Mount("/me", app.meHandler())
	root.Mount("/logout", app.logoutHandler())

	err = http.ListenAndServe(":8080", root)
	if err != nil {
		logger.Err(err).Msg("couldn't start server")
	}
}

func (app *App) lastUsernamesHandler() http.Handler {
	router := chi.NewRouter()
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		usernames, err := app.storage.LastUsernames()
		if err != nil {
			app.logger.Error().Err(err).Msg("failed lastUsernamesHandler")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := app.Templates.LastUsernames.Execute(w, usernames); err != nil {
			app.logger.Error().Err(err).Msg("failed to render last usernames")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	return router
}

func (app *App) logoutHandler() http.Handler {
	router := chi.NewRouter()
	router.Use(app.checkAuthedAndRedirect(false, "/"))
	router.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		token := GetToken(r.Context())
		if err := app.storage.DeleteToken(token); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		RemoveAuthCookie(rw)
		redirect(rw, r, "/")
	})
	return router
}

func (app *App) meHandler() http.Handler {
	router := chi.NewRouter()
	router.Use(app.checkAuthedAndRedirect(false, "/login"))
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		if err := app.Templates.User.Execute(w, user.ToUserInfo(true)); err != nil {
			app.logger.Error().Err(err).Msg("failed to render user page")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	return router
}

func (app *App) usersHandler() http.Handler {
	router := chi.NewRouter()
	router.Get("/{username}", func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		if username == "" {
			w.Write([]byte("username not found"))
			return
		}
		user, err := app.storage.FindUserByUsername(username)
		if err != nil {
			app.logger.Error().Err(err).Msg("failed to get user by username")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if user == nil {
			w.Write([]byte("username not found"))
			return
		}
		if err := app.Templates.User.Execute(w, user.ToUserInfo(false)); err != nil {
			app.logger.Error().Err(err).Msg("failed to render user page")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	return router
}

func (app *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())
	if user != nil {
		app.logger.Info().Str("username", user.Username).
			Msg("request for user index page, redirecting")
		redirect(w, r, "/user/"+user.Username)
		return
	}
	if err := app.Templates.Index.Execute(w, nil); err != nil {
		app.logger.Error().Err(err).Msg("failed to render index page")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (app *App) loginHandler() http.Handler {
	loginRouter := chi.NewRouter()
	loginRouter.Use(app.checkAuthedAndRedirect(true, "/me"))
	loginRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			app.logger.Error().Err(err).Msg("failed to parse url")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		infoText := r.Form.Get("info")

		loginInfo := templates.LoginInfo{
			Username: "",
			Password: "",
			Hint: templates.Hint{
				HintText: infoText,
				IsError:  false,
			},
		}
		if err := app.Templates.Login.Execute(w, &loginInfo); err != nil {
			app.logger.Error().Err(err).Msg("failed to render template")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	loginRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			app.logger.Error().Err(err).Msg("failed to parse form")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		loginInfo := templates.NewLoginInfo(r.Form)
		if loginInfo.Username == "" {
			hint := templates.Hint{
				HintText: "username is empty",
				IsError:  true,
			}
			app.respondLoginHint(w, loginInfo, http.StatusBadRequest, hint)
			return
		}
		usr, err := app.storage.FindUserByUsername(loginInfo.Username)
		if err != nil {
			app.logger.Error().Err(err).Msg("failed storage username search")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if usr == nil || usr.PasswordHash != model.HashPassword(loginInfo.Password) {
			hint := templates.Hint{
				HintText: "user not found or password is wrong",
				IsError:  true,
			}
			app.respondLoginHint(w, loginInfo, http.StatusUnauthorized, hint)
			return
		}

		newToken := uuid.NewV1()
		app.logger.Info().Str("userID", usr.ID.String()).Msg("user logged in, generated new token")
		if err := app.storage.InsertToken(newToken, usr.ID); err != nil {
			app.logger.Error().Err(err).Msg("failed to insert new token")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		SetAuthCookie(w, newToken)

		redirect(w, r, "/")
	})

	return loginRouter
}

func (app *App) respondLoginHint(w http.ResponseWriter, loginInfo *templates.LoginInfo,
	status int, hint templates.Hint) {
	loginInfo.ToResponse(hint)
	w.WriteHeader(status)
	if err := app.Templates.Login.Execute(w, loginInfo); err != nil {
		app.logger.Error().Err(err).Msg("failed to render template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (app *App) signupHandler() http.Handler {
	signupRouter := chi.NewRouter()
	signupRouter.Use(app.checkAuthedAndRedirect(true, "/me"))
	signupRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		defaultResponse := templates.SignupInfo{
			Gender: "other",
		}
		if err := app.Templates.Signup.Execute(w, &defaultResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	signupRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			app.logger.Err(err).Msg("failed parsing form")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		info := templates.NewSignupInfo(r.Form)
		app.logger.Info().Msgf("parsed form: %+v", *info)

		usr, err := model.NewUserFromSignup(info)
		if err != nil {
			app.returnErrorOnSingUp(w, info, err)
			return
		}
		anotherUsr, err := app.storage.FindUserByUsername(usr.Username)
		if err != nil {
			app.logger.Err(err).Msg("failed to check for existing username in storage")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if anotherUsr != nil {
			app.returnErrorOnSingUp(w, info, errors.New("username already exists, choose another one"))
			return
		}
		if err = app.storage.InsertUser(usr); err != nil {
			app.logger.Err(err).Msg("failed to store new user")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		redirect(w, r, "/login?info=registeredOK")
	})
	return signupRouter
}

func (app *App) returnErrorOnSingUp(w http.ResponseWriter, info *templates.SignupInfo, err error) {
	app.logger.Info().Err(err).Msg("falied to sing up")
	info.Err = err.Error()
	info.Password = ""
	if err := app.Templates.Signup.Execute(w, info); err != nil {
		app.logger.Err(err).Msg("failed to render template")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type LogEntry struct {
	logger *zerolog.Logger
	r      *http.Request
}

func (le LogEntry) Write(status, bytes int, elapsed time.Duration) {
	le.logger.Info().
		Str("requestID", middleware.GetReqID(le.r.Context())).
		Str("url", le.r.URL.String()).
		Str("method", le.r.Method).
		Int("status", status).
		Int64("duration", elapsed.Milliseconds()).
		Msg("request done")
}

func (le LogEntry) Panic(v interface{}, stack []byte) {
	le.logger.Panic().
		Str("panic", fmt.Sprintf("%v", v)).
		Str("stack", string(stack)).
		Msg("panic!")
}

type RequestFormatter struct {
	*zerolog.Logger
}

func (rf RequestFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return LogEntry{r: r, logger: rf.Logger}
}

func redirect(w http.ResponseWriter, r *http.Request, where string) {
	http.Redirect(w, r, where, http.StatusSeeOther)
}
