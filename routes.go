package main

import (
	"fmt"
	"net/http"
	"strconv"
	"testapp/data"

	"github.com/go-chi/chi/v5"
	"github.com/xsdrt/hispeed2/mailer"
)

func (a *application) routes() *chi.Mux {
	// middleware must come before any routes chi router expects to find middleware 1st so always at the top...

	// add routes here...
	a.get("/", a.Handlers.Home)
	a.App.Routes.Get("/go-page", a.Handlers.GoPage)
	a.App.Routes.Get("/jet-page", a.Handlers.JetPage)
	a.App.Routes.Get("/sessions", a.Handlers.SessionTest)

	a.App.Routes.Get("/users/login", a.Handlers.UserLogin)
	a.post("/users/login", a.Handlers.PostUserLogin)
	a.App.Routes.Get("/users/logout", a.Handlers.Logout)

	a.App.Routes.Get("/form", a.Handlers.Form)
	a.App.Routes.Post("/form", a.Handlers.PostForm)

	a.get("/json", a.Handlers.JSON)
	a.get("/xml", a.Handlers.XML)
	a.get("/download-file", a.Handlers.DownloadFile)

	a.get("/crypto", a.Handlers.TestCrypto)

	a.get("/cache-test", a.Handlers.ShowCachePage)
	a.post("/api/save-in-cache", a.Handlers.SaveInCache)
	a.post("/api/get-from-cache", a.Handlers.GetFromCache)
	a.post("/api/delete-from-cache", a.Handlers.DeleteFromCache)
	a.post("/api/empty-cache", a.Handlers.EmptyCache)

	a.get("/test-mail", func(w http.ResponseWriter, r *http.Request) {
		msg := mailer.Message{
			From:        "test@example.com",
			To:          "m_redinger@hotmail.com",
			Subject:     "Test Subject - sent using an channel",
			Template:    "test",
			Attachments: nil,
			Data:        nil,
		}

		// a.App.Mail.Jobs <- msg
		// res := <-a.App.Mail.Results
		// if res.Error != nil {
		// 	a.App.ErrorLog.Println(res.Error)
		// }
		err := a.App.Mail.SendSTMPMessage(msg)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		fmt.Fprint(w, "Sent mail!")
	})

	a.App.Routes.Get("/create-user", func(w http.ResponseWriter, r *http.Request) {
		u := data.User{
			FirstName: "Michael",
			LastName:  "Redinger",
			Email:     "mia@somewhere.com",
			Active:    1,
			Password:  "password",
		}

		id, err := a.Models.Users.Insert(u)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		fmt.Fprintf(w, "%d: %s", id, u.FirstName)
	})

	a.App.Routes.Get("/get-all-users", func(w http.ResponseWriter, r *http.Request) { // inline test func...
		users, err := a.Models.Users.GetAll()
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}
		for _, x := range users {
			fmt.Fprint(w, x.LastName)
		}
	})

	a.App.Routes.Get("/get-user/{id}", func(w http.ResponseWriter, r *http.Request) { // inline test func...
		id, _ := strconv.Atoi(chi.URLParam(r, "id")) // gives the id thats in the url...

		u, err := a.Models.Users.Get(id)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}
		fmt.Fprintf(w, "%s %s %s", u.FirstName, u.LastName, u.Email)
	})

	a.App.Routes.Get("/update-user/{id}", func(w http.ResponseWriter, r *http.Request) { // inline test func...
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		u, err := a.Models.Users.Get(id)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		u.LastName = a.App.RandomString(10)

		// Testing validation...
		validator := a.App.Validator(nil)
		//validator.Check(len(u.LastName) > 20, "last_name", "Last name must be 20 characters or more") // obviousley want this to fail...
		u.LastName = ""

		u.Validate(validator)

		if !validator.Valid() {
			fmt.Fprint(w, "failed validation") // This should print out on a empty web page when we try and update a last name less than 20 characters...
			return
		}

		err = u.Update(*u)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		fmt.Fprintf(w, "updated last name to %s", u.LastName)
	})

	// static routes
	fileServer := http.FileServer(http.Dir("./public"))
	a.App.Routes.Handle("/public/*", http.StripPrefix("/public", fileServer))

	return a.App.Routes
}
