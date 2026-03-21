package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	mail "github.com/xhit/go-simple-mail/v2"
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

	mux.Get("/test-email", func(w http.ResponseWriter, r *http.Request) {

		m := Mail{
			Domain:      "localhost",
			Host:        "localhost",
			Port:        1025,
			Username:    "",
			Password:    "",
			Encryption:  mail.EncryptionNone,
			FromAddress: "no-reply@example.com",
			FromName:    "Example App",
		}

		msg := Message{
			To:       "test@example.com",
			Subject:  "Test Email",
			Template: "mail",
			Data:     "MailHog test email from subscription_service",
		}

		if err := m.SendMail(msg); err != nil {
			app.ErrorLog.Printf("send test email: %v", err)
			http.Error(w, "unable to send test email: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test email sent to MailHog; open http://localhost:8025"))
	})

	return mux
}
