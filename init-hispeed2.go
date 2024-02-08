package main

import (
	"log"
	"os"
	"testapp/handlers"

	"github.com/xsdrt/hispeed2"
)

func initApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err) //if can't find the wd then just die as something went wrong...
	}

	// init hispeed2
	his := &hispeed2.HiSpeed2{}
	err = his.New(path)
	if err != nil {
		log.Fatal(err) // again something went wrong , so just die...
	}

	his.AppName = "testapp"

	myHandlers := &handlers.Handlers{
		App: his,
	}

	app := &application{
		App:      his,
		Handlers: myHandlers,
	}

	app.App.Routes = app.routes()

	return app

}
