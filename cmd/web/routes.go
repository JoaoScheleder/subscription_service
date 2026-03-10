package main

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(app.SessionLoad)

	mux.Get("/", app.HomePage)

	mux.Get("/login", app.LoginPage)
	mux.Post("/login", app.PostLoginPage)

	mux.Get("/logout", app.LogoutPage)

	mux.Get("/register", app.RegisterPage)
	mux.Post("/register", app.PostRegisterPage)

	mux.Get("/activate", app.ActivateAccount)

	return mux
}
