package main

import "net/http"

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
		app.Session.Put(r.Context(), "error", "Invalid Credentials.")
		app.ErrorLog.Printf("get user by email: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	validPassword, err := user.PasswordMatches(password)
	if err != nil || !validPassword {
		app.Session.Put(r.Context(), "error", "Invalid Credentials.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "userID", user.ID)
	app.Session.Put(r.Context(), "User", user)

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (app *Config) LogoutPage(w http.ResponseWriter, r *http.Request) {

}

func (app *Config) RegisterPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "register.page.gohtml", nil)
}

func (app *Config) PostRegisterPage(w http.ResponseWriter, r *http.Request) {

}

func (app *Config) ActivateAccount(w http.ResponseWriter, r *http.Request) {

}
