package main

import (
	"errors"
	"fmt"
	"net/http"
	data "subscription_service/models"

	"github.com/jackc/pgx/v5"
)

func (app *Config) HomePage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.gohtml", nil)
}

func (app *Config) LoginPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.gohtml", nil)
}

func (app *Config) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	_ = app.Session.RenewToken(r.Context())

	// parse form post
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Printf("parse form: %v", err)
		http.Error(w, "unable to parse form", http.StatusBadRequest)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	if email == "" || password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			app.Session.Put(r.Context(), "error", "Invalid Credentials.")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		app.Session.Put(r.Context(), "error", "Invalid Credentials.")
		app.ErrorLog.Printf("get user by email: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	validPassword, err := user.PasswordMatches(password)
	if err != nil || !validPassword {
		if !validPassword {
			msg := Message{
				To:      user.Email,
				Subject: "Failed Login Attempt",
				Data:    "There was a failed login attempt on your account. If this wasn't you, please reset your password immediately.",
			}
			app.sendEmail(msg)
		}
		app.Session.Put(r.Context(), "error", "Invalid Credentials.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "userID", user.ID)
	app.Session.Put(r.Context(), "User", user)

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (app *Config) LogoutPage(w http.ResponseWriter, r *http.Request) {
	err := app.Session.Destroy(r.Context())
	if err != nil {
		app.ErrorLog.Printf("destroy session: %v", err)
		http.Error(w, "unable to logout", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *Config) RegisterPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "register.page.gohtml", nil)
}

func (app *Config) PostRegisterPage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ErrorLog.Printf("parse form: %v", err)
		http.Error(w, "unable to parse form", http.StatusBadRequest)
		return
	}

	u := data.User{
		Email:     r.Form.Get("email"),
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Password:  r.Form.Get("password"),
		Active:    0,
		IsAdmin:   0,
	}

	_, err = u.Insert(u)

	if err != nil {
		app.Session.Put(r.Context(), "error", "Unable to create account. Please try again.")
		app.ErrorLog.Printf("insert user: %v", err)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	url := fmt.Sprintf("http://%s/activate?email=%s", r.Host, u.Email)
	signedUrl := GenerateTokenFromString(url)

	app.InfoLog.Printf("activation URL: %s", signedUrl)

	msg := Message{
		To:      u.Email,
		Subject: "Activate Your Account",
		Data:    fmt.Sprintf("Please click the following link to activate your account: %s", signedUrl),
	}

	app.sendEmail(msg)

	app.Session.Put(r.Context(), "flash", "Registration successful! Please check your email to activate your account.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)

}

func (app *Config) ActivateAccount(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	hash := r.URL.Query().Get("hash")

	if email == "" || hash == "" {
		http.Error(w, "invalid activation link", http.StatusBadRequest)
		return
	}

	if !VerifyToken(hash) {
		http.Error(w, "invalid or expired activation link", http.StatusBadRequest)
		return
	}

	user, err := app.Models.User.GetByEmail(email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "invalid activation link", http.StatusBadRequest)
			return
		}
		app.ErrorLog.Printf("get user by email: %v", err)
		http.Error(w, "unable to activate account", http.StatusInternalServerError)
		return
	}

	user.Active = 1

	err = user.Update()
	if err != nil {
		app.ErrorLog.Printf("update user: %v", err)
		http.Error(w, "unable to activate account", http.StatusInternalServerError)
		return
	}

	app.Session.Put(r.Context(), "flash", "Account activated! You can now log in.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
