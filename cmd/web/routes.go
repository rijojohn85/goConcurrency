package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() http.Handler {
	// create router
	mux := chi.NewRouter()

	// set up middleware

	mux.Use(middleware.Recoverer)
	mux.Use(app.SessionLoad)

	// define app routes

	mux.Get("/", app.HomePage)

	mux.Get("/login", app.LoginPage)
	mux.Post("/login", app.PostLoginPage)

	mux.Get("/logout", app.Logout)
	mux.Get("/register", app.RegisterPage)
	mux.Post("/register", app.PostRegisterPage)
	mux.Get("/activate", app.ActivateAccount)

	return mux
}

// mux.Get("/test-email", func(w http.ResponseWriter, r *http.Request) {
// 	m := Mail{
// 		Domain:      "localhost",
// 		Host:        "localhost",
// 		Port:        1025,
// 		Encryption:  "none",
// 		FromName:    "My Company",
// 		FromAddress: "info@mycompany.com",
// 		ErrorChan:   make(chan error),
// 	}
// 	msg := Message{
// 		To:      "me@here.com",
// 		Subject: "Test Email",
// 		Data:    "Hello, World!",
// 	}
// 	m.sendMail(msg, make(chan error))
// })
