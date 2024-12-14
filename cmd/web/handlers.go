package main

import (
	"net/http"
)

func (app *Config) HomePage(w http.ResponseWriter, req *http.Request) {
	app.render(w, req, "home.page.gohtml", nil)
}

func (app *Config) LoginPage(w http.ResponseWriter, req *http.Request) {
	app.render(w, req, "login.page.gohtml", nil)
}

func (app *Config) PostLoginPage(w http.ResponseWriter, req *http.Request) {
	_ = app.Session.RenewToken(req.Context())
	// parse form post
	err := req.ParseForm()
	if err != nil {
		app.ErrorLog.Println(err, "getform")
		app.Session.Put(req.Context(), "error", "get form invalid credeintals")
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}
	// get email and passowrd
	email := req.Form.Get("email")
	password := req.Form.Get("password")
	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		app.ErrorLog.Println(err, "get datab")
		app.Session.Put(req.Context(), "error", "get user invalid credeintals")
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}
	passCheck, err := user.PasswordMatches(password)
	if !passCheck || err != nil {
		if !passCheck {
			msg := Message{
				To:      email,
				Subject: "Failed Login Attempt",
				Data:    "Invalid Login Attempt",
			}
			app.sendEmail(msg)
		}
		app.ErrorLog.Println(err, "get passCheck", password)
		app.Session.Put(req.Context(), "error", "invalid credeintals")
		http.Redirect(w, req, "/login", http.StatusUnauthorized)
		return
	}
	app.Session.Put(req.Context(), "userID", user.ID)
	app.Session.Put(req.Context(), "user", user)
	app.Session.Put(req.Context(), "flash", "Successful login")
	// redirect the user
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func (app *Config) Logout(w http.ResponseWriter, req *http.Request) {
	_ = app.Session.Destroy(req.Context())
	_ = app.Session.RenewToken(req.Context())
	app.Session.Put(req.Context(), "flash", "Successful logout")
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func (app *Config) RegisterPage(w http.ResponseWriter, req *http.Request) {
	app.render(w, req, "register.page.gohtml", nil)
}

func (app *Config) PostRegisterPage(w http.ResponseWriter, req *http.Request) {
}

func (app *Config) ActivateAccount(w http.ResponseWriter, req *http.Request) {
}
