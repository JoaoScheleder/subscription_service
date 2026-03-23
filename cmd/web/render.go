package main

import (
	"fmt"
	"net/http"
	data "subscription_service/models"
	"time"

	"html/template"
)

const PATH_TO_TEMPLATES = "./cmd/web/templates"

type TemplateData struct {
	StringMap     map[string]string
	IntMap        map[string]int
	FloatMap      map[string]float64
	Data          map[string]any
	Flash         string
	Warning       string
	Error         string
	Authenticated bool
	Now           time.Time
	User          *data.User
}

func (app *Config) render(w http.ResponseWriter, r *http.Request, tmpl string, templateData *TemplateData) {
	if templateData == nil {
		templateData = &TemplateData{}
	}
	templateData = app.AddDefaultData(templateData, r)

	partials := []string{
		fmt.Sprintf("%s/base.layout.gohtml", PATH_TO_TEMPLATES),
		fmt.Sprintf("%s/header.partial.gohtml", PATH_TO_TEMPLATES),
		fmt.Sprintf("%s/navbar.partial.gohtml", PATH_TO_TEMPLATES),
		fmt.Sprintf("%s/footer.partial.gohtml", PATH_TO_TEMPLATES),
		fmt.Sprintf("%s/alerts.partial.gohtml", PATH_TO_TEMPLATES),
	}

	var templateSlice []string
	templateSlice = append(templateSlice, fmt.Sprintf("%s/%s", PATH_TO_TEMPLATES, tmpl))

	templateSlice = append(templateSlice, partials...)

	parsedTemplate, err := template.ParseFiles(templateSlice...)
	if err != nil {
		app.ErrorLog.Printf("parse template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = parsedTemplate.Execute(w, templateData)
	if err != nil {
		app.ErrorLog.Printf("execute template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *Config) AddDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	if app.IsAuthenticatedUser(r) {
		td.Authenticated = true
		user, ok := app.Session.Get(r.Context(), "user").(data.User)
		if !ok {
			app.ErrorLog.Printf("error getting user from session")
			return td
		} else {
			td.User = &user
		}
	}
	td.Now = time.Now()

	return td
}

func (app *Config) IsAuthenticatedUser(r *http.Request) bool {
	return app.Session.Exists(r.Context(), "userID")
}
