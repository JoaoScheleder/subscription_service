package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	data "subscription_service/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
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

func (app *Config) ChooseSubscription(w http.ResponseWriter, r *http.Request) {

	plans, err := app.Models.Plan.GetAll()
	if err != nil {
		app.ErrorLog.Printf("get all plans: %v", err)
		http.Error(w, "unable to load subscription plans", http.StatusInternalServerError)
		return
	}

	data := make(map[string]any)
	data["plans"] = plans

	app.render(w, r, "plans.page.gohtml", &TemplateData{Data: data})

}

func (app *Config) SubscribeToPlan(w http.ResponseWriter, r *http.Request) {

	// get the plan id
	planID := r.URL.Query().Get("id")
	if planID == "" {
		http.Error(w, "missing plan id", http.StatusBadRequest)
		return
	}

	planIDInt, err := strconv.Atoi(planID)

	// get the plan
	plan, err := app.Models.Plan.GetOne(planIDInt)
	if err != nil {
		app.ErrorLog.Printf("get plan by id: %v", err)
		http.Error(w, "unable to get plan", http.StatusInternalServerError)
		return
	}

	// get the user
	user, ok := app.Session.Get(r.Context(), "user").(data.User)
	if !ok {
		app.ErrorLog.Printf("error getting user from session")
		http.Error(w, "unable to get user", http.StatusInternalServerError)
		return
	}

	// generate an invoice

	app.WaitGroup.Add(1)
	go func() {
		defer app.WaitGroup.Done()
		invoice, err := app.getInvoice(user, plan)
		if err != nil {
			app.ErrorLog.Printf("create invoice: %v", err)
			app.ErrorChan <- err
			return
		}

		msg := Message{
			To:       user.Email,
			Subject:  "Your Subscription Invoice",
			Data:     invoice,
			Template: "invoice",
		}

		app.sendEmail(msg)
	}()

	app.WaitGroup.Add(1)
	go func() {
		defer app.WaitGroup.Done()

		pdf := app.GenerateManual(user, plan)
		err := pdf.OutputFileAndClose(fmt.Sprintf("./tmp/%d_manual.pdf", user.ID))
		if err != nil {
			app.ErrorLog.Printf("generate manual: %v", err)
			app.ErrorChan <- err
			return
		}

		msg := Message{
			To:          user.Email,
			Subject:     "Your Subscription Manual",
			Data:        "Please find your subscription manual attached.",
			Template:    "manual",
			Attachments: []string{fmt.Sprintf("./tmp/%d_manual.pdf", user.ID)},
		}

		app.sendEmail(msg)

	}()

	app.Session.Put(r.Context(), "flash", fmt.Sprintf("You have subscribed to the %s plan! Please check your email for your invoice and manual.", plan.PlanName))
	http.Redirect(w, r, "/memebers/plans", http.StatusSeeOther)
}

func (app *Config) GenerateManual(user data.User, plan *data.Plan) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(10, 13, 10)

	importer := gofpdi.NewImporter()

	time.Sleep(5 * time.Second)

	tpl := importer.ImportPage(pdf, "./pdf/manual.pdf", 1, "/MediaBox")
	pdf.AddPage()
	importer.UseImportedTemplate(pdf, tpl, 0, 0, 210, 0)

	pdf.SetX(75)
	pdf.SetY(150)

	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 4, fmt.Sprintf("%s %s", user.FirstName, user.LastName), "", "C", false)

	pdf.Ln(5)

	pdf.MultiCell(0, 4, fmt.Sprintf("%s User Guide", plan.PlanName), "", "C", false)

	return pdf
}

func (app *Config) getInvoice(user data.User, plan *data.Plan) (string, error) {
	// formatte the plan amount in dollars
	amount := fmt.Sprintf("%.2f", float64(plan.PlanAmount)/100)

	// create the invoice
	invoice := fmt.Sprintf("Invoice for %s %s\n\nPlan: %s\nAmount: $%s\n\nThank you for your subscription!", user.FirstName, user.LastName, plan.PlanName, amount)
	return invoice, nil
}
